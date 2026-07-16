// Пример: middleware в net/http.
//
// Middleware — это функция-обёртка вокруг http.Handler:
//
//	type Middleware func(http.Handler) http.Handler
//
// Она может сделать что-то ДО вызова next, передать управление дальше,
// а затем (опционально) сделать что-то ПОСЛЕ — например, залогировать ответ.
//
// Запуск:
//
//	go run ./Chapter4/examples/02_middleware/
//
// Проверка:
//
//	curl -i http://localhost:8081/ok
//	curl -i http://localhost:8081/panic
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"
)

// ---------------------------------------------------------------------------
// Context key
// ---------------------------------------------------------------------------

// ctxKey — отдельный тип для ключей context.WithValue.
// Так мы не пересечёмся со строковыми ключами из других пакетов
// (collision-safe pattern из стандартной библиотеки).
type ctxKey int

// keyRequestID — ключ, под которым в context лежит request ID.
// Значение 1 ни на что не влияет: важен сам тип+константа как уникальный ключ.
const keyRequestID ctxKey = 1

// ---------------------------------------------------------------------------
// Middleware type + Chain
// ---------------------------------------------------------------------------

// Middleware принимает «внутренний» handler и возвращает новый handler,
// который оборачивает next. Это и есть весь паттерн — чистая композиция функций.
type Middleware func(http.Handler) http.Handler

// Chain собирает несколько middleware в одну обёртку вокруг h.
//
// Важно: цикл идёт С КОНЦА списка к началу.
// Если вызвать Chain(mux, A, B, C), получится:
//
//	A(B(C(mux)))
//
// То есть при запросе порядок такой:
//
//	Request  → A → B → C → mux
//	Response ← A ← B ← C ← mux
//
// Первый в списке — самый «внешний»: он видит запрос первым и ответ последним.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// ---------------------------------------------------------------------------
// RequestID — прокидывает / генерирует X-Request-Id
// ---------------------------------------------------------------------------

// RequestID:
//  1. Берёт X-Request-Id из входящего запроса (удобно для трейсинга от клиента/прокси).
//  2. Если заголовка нет — генерирует новый id.
//  3. Кладёт id в request context (чтобы handler и другие middleware могли его читать).
//  4. Дублирует id в ответный заголовок X-Request-Id.
//
// Context здесь — место для request-scoped данных на время одного HTTP-запроса.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = newID()
		}

		// context.WithValue возвращает НОВЫЙ context; исходный r.Context() не меняется.
		ctx := context.WithValue(r.Context(), keyRequestID, id)

		// Клиент/прокси увидят тот же id в ответе — удобно склеивать логи.
		w.Header().Set("X-Request-Id", id)

		// r.WithContext(ctx) — новый *http.Request с нашим context.
		// Именно его передаём дальше по цепочке.
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ---------------------------------------------------------------------------
// statusWriter — обёртка над ResponseWriter, чтобы узнать status code
// ---------------------------------------------------------------------------

// statusWriter нужен, потому что стандартный http.ResponseWriter
// не отдаёт наружу код статуса после WriteHeader.
// Logging middleware оборачивает w в statusWriter и читает sw.status после handler.
//
// Встраивание http.ResponseWriter (embedding) пробрасывает все остальные методы
// (Header, Write, …) на оригинальный writer автоматически.
type statusWriter struct {
	http.ResponseWriter
	status int // код, который реально отправили клиенту
}

// WriteHeader перехватывает вызов handler'а: запоминаем code и делегируем дальше.
// Если handler так и не вызвал WriteHeader, в Logging останется StatusOK (200) —
// это совпадает с поведением net/http: первый Write() неявно шлёт 200.
func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// ---------------------------------------------------------------------------
// Logging — method, path, status, duration, request id
// ---------------------------------------------------------------------------

// Logging замеряет время обработки и пишет одну строку лога после next.
// Status берём из statusWriter; request id — из context (его положил RequestID).
//
// Порядок в Chain важен: RequestID должен быть СНАРУЖИ Logging,
// иначе к моменту лога id ещё не будет в context.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// По умолчанию 200: если handler только пишет тело без WriteHeader.
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		// Передаём sw вместо w — handler пишет в наш wrapper.
		next.ServeHTTP(sw, r)

		// Type assertion: Value возвращает any; ожидаем string.
		// Если ключа нет — id будет "" (второй результат ok игнорируем).
		id, _ := r.Context().Value(keyRequestID).(string)

		log.Printf("%s %s %d %s id=%s", r.Method, r.URL.Path, sw.status, time.Since(start), id)
	})
}

// ---------------------------------------------------------------------------
// Recover — не даём panic уронить весь процесс
// ---------------------------------------------------------------------------

// Recover ловит panic в handler (и во внутренних middleware) через recover().
// Без этого один упавший запрос может завершить весь HTTP-сервер.
//
// defer + recover должны стоять в той же горутине, где случился panic
// (ServeHTTP вызывается синхронно — это ок).
//
// Обычно Recover ставят ближе к handler (внутри), а Logging — снаружи,
// чтобы в лог попал и 500 после паники. В этом примере порядок:
// RequestID → Logging → Recover → mux.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic: %v", rec)
				// Клиенту не отдаём детали паники — только общий 500.
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// newID генерирует короткий случайный hex-id (8 байт → 16 hex-символов).
// Для учебного примера достаточно; в проде часто берут UUID.
func newID() string {
	var b [8]byte
	_, _ = rand.Read(b[:]) // игнорируем ошибку: для демо достаточно
	return hex.EncodeToString(b[:])
}

// ---------------------------------------------------------------------------
// main — маршруты и сборка цепочки
// ---------------------------------------------------------------------------

func main() {
	mux := http.NewServeMux()

	// Успешный handler: читает request id из context и пишет его в ответ.
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) {
		id, _ := r.Context().Value(keyRequestID).(string)
		fmt.Fprintf(w, "ok request_id=%s\n", id)
	})

	// Handler специально паникует — чтобы показать работу Recover.
	mux.HandleFunc("GET /panic", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// Собираем цепочку. Порядок аргументов = порядок на пути запроса:
	//
	//   Request  → RequestID → Logging → Recover → mux
	//   Response ← RequestID ← Logging ← Recover ← mux
	//
	// RequestID снаружи — id доступен и Logging, и handler'ам.
	// Logging снаружи Recover — после паники всё равно залогируем duration/status.
	// Recover ближе к mux — ловит panic в handler до того, как он уйдёт наружу.
	handler := Chain(mux, RequestID, Logging, Recover)

	log.Println("listening on :8081 — try GET /ok and GET /panic")
	log.Fatal(http.ListenAndServe(":8081", handler))
}
