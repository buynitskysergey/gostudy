package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

//go:embed openapi.yaml
var openAPISpec []byte

//go:embed swagger.html
var swaggerUI []byte

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(openAPISpec)
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(swaggerUI)
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
			return
		}
		title := strings.TrimSpace(req.Title)
		if title == "" {
			http.Error(w, `{"error":"title required"}`, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":    "demo-1",
			"title": title,
			"done":  false,
		})
	})

	log.Println("listening on :8081")
	log.Println("  GET  /docs          — Swagger UI")
	log.Println("  GET  /openapi.yaml  — contract")
	log.Println("  POST /api/v1/tasks  — implements the contract")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
