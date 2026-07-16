package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type CreateTaskRequest struct {
	Title string `json:"title"`
}

type validationError struct {
	fields map[string]string
}

func (e validationError) Error() string { return "validation failed" }

func (e validationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"error":  "validation failed",
		"fields": e.fields,
	})
}

func (r CreateTaskRequest) Validate() error {
	fields := map[string]string{}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		fields["title"] = "required"
	} else if len(title) > 200 {
		fields["title"] = "max length is 200"
	}
	if len(fields) > 0 {
		return validationError{fields: fields}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var req CreateTaskRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
			return
		}
		if err := req.Validate(); err != nil {
			writeJSON(w, http.StatusBadRequest, err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{
			"title": strings.TrimSpace(req.Title),
		})
	})

	log.Println("listening on :8081 — POST /tasks")
	log.Println(`  {"title":""}           → 400 fields.title=required`)
	log.Println(`  {"title":"buy milk"}   → 201`)
	log.Fatal(http.ListenAndServe(":8081", mux))
}
