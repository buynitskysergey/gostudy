package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "file:ch5_04.db?cache=shared&mode=rwc")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Узкий пул: одновременно «в БД» могут быть только 2 goroutine.
	const maxOpen = 2
	const workers = 8
	const work = 2 * time.Second

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)

	ctx := context.Background()
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS ping(x INTEGER); INSERT OR IGNORE INTO ping(x) VALUES (1)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("MaxOpenConns=%d, workers=%d, work=%s на каждый запрос\n", maxOpen, workers, work)
	fmt.Printf("ожидание: ~%d волн × %s = ~%s (только %d conn одновременно)\n\n",
		workers/maxOpen, work, time.Duration(workers/maxOpen)*work, maxOpen)

	var wg sync.WaitGroup
	start := time.Now()
	elapsed := func() string {
		return time.Since(start).Round(time.Millisecond).String()
	}

	for i := 1; i <= workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			fmt.Printf("[%s] worker %d: ждёт conn из пула...\n", elapsed(), id)

			// Conn() занимает слот пула до Close — имитация долгого хендлера,
			// который держит соединение на время «работы» (как незакрытый Rows/Tx).
			cctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			conn, err := db.Conn(cctx)
			if err != nil {
				log.Printf("[%s] worker %d: %v", elapsed(), id, err)
				return
			}

			st := db.Stats()
			fmt.Printf("[%s] worker %d: взял conn (inUse=%d wait=%d) → работа %s\n",
				elapsed(), id, st.InUse, st.WaitCount, work)

			var x int
			if err := conn.QueryRowContext(cctx, `SELECT x FROM ping`).Scan(&x); err != nil {
				_ = conn.Close()
				log.Printf("[%s] worker %d: %v", elapsed(), id, err)
				return
			}
			time.Sleep(work)

			_ = conn.Close() // вернули conn в пул — следующий воркер может стартовать
			fmt.Printf("[%s] worker %d: готово, conn возвращён\n", elapsed(), id)
		}(i)
	}

	wg.Wait()
	st := db.Stats()
	took := time.Since(start).Round(time.Millisecond)
	fmt.Printf("\nfinished in %s (не ~%s: пул сериализовал %d воркеров в %d слота)\n",
		took, work, workers, maxOpen)
	fmt.Printf("db.Stats: Open=%d InUse=%d WaitCount=%d WaitDuration=%s\n",
		st.OpenConnections, st.InUse, st.WaitCount, st.WaitDuration.Round(time.Millisecond))
}
