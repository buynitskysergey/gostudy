package task

import (
	"errors"
	"net/http"

	"study1/Chapter4/examples/app/internal/httpapi"
)

type Handler struct {
	store        *Store
	maxBodyBytes int64
}

func NewHandler(store *Store, maxBodyBytes int64) *Handler {
	return &Handler{store: store, maxBodyBytes: maxBodyBytes}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/tasks", h.list)
	mux.HandleFunc("POST /api/v1/tasks", h.create)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.get)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.delete)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	httpapi.WriteJSON(w, http.StatusOK, h.store.List())
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := httpapi.DecodeJSON(r, &req, h.maxBodyBytes); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if err := req.Validate(); err != nil {
		var ve ValidationError
		if errors.As(err, &ve) {
			httpapi.WriteJSON(w, http.StatusBadRequest, ve)
			return
		}
		httpapi.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	created := h.store.Create(req.NormalizedTitle())
	httpapi.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	t, err := h.store.Get(id)
	if errors.Is(err, ErrNotFound) {
		httpapi.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, t)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.store.Delete(id); errors.Is(err, ErrNotFound) {
		httpapi.WriteError(w, http.StatusNotFound, "task not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
