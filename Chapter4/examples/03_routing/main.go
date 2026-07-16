package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
)

type store struct {
	mu    sync.RWMutex
	items map[string]string
}

func main() {
	s := &store{items: map[string]string{"1": "alpha", "2": "beta"}}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		fmt.Fprintf(w, "count=%d\n", len(s.items))
		ids := make([]string, 0, len(s.items))
		for id := range s.items {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fmt.Fprintf(w, "%s %s\n", id, s.items[id])
		}
	})
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		s.mu.RLock()
		name, ok := s.items[id]
		s.mu.RUnlock()
		if !ok {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "%s=%s\n", id, name)
	})
	mux.HandleFunc("PUT /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "name query required", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.items[id] = name
		log.Println("item added", id, name)
		s.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})

	log.Println("listening on :8083")
	log.Println("  GET  /items")
	log.Println("  GET  /items/{id}")
	log.Println("  PUT  /items/{id}?name=...")
	log.Println("  POST /items  → 405 (only GET registered for /items)")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
