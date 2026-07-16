package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorBody struct {
	Error string `json:"error"`
}

func decodeJSON(r *http.Request, dst any, maxBytes int64) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode: %v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Title string `json:"title"`
		}
		if err := decodeJSON(r, &in, 1<<20); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorBody{Error: err.Error()})
			return
		}
		out := Task{
			ID:        "t-1",
			Title:     in.Title,
			Done:      false,
			CreatedAt: time.Now().UTC(),
		}
		writeJSON(w, http.StatusOK, out)
	})

	log.Println("listening on :8081 — POST /echo with JSON {\"title\":\"...\"}")
	log.Println(`try unknown field: {"title":"x","extra":1} → 400`)
	fmt.Println()
	log.Fatal(http.ListenAndServe(":8081", mux))
}
