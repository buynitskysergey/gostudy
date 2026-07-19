package account

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"study1/Chapter5/examples/app/internal/httpapi"
)
ё
type Service interface {
	Create(ctx context.Context, owner string, balance int64) (Account, error)
	Get(ctx context.Context, id int64) (Account, error)
	Transfer(ctx context.Context, key string, req TransferRequest) (IdempotentResult, error)
}

type Handler struct {
	svc          Service
	maxBodyBytes int64
}

func NewHandler(svc Service, maxBodyBytes int64) *Handler {
	return &Handler{svc: svc, maxBodyBytes: maxBodyBytes}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/accounts", h.create)
	mux.HandleFunc("GET /api/v1/accounts/{id}", h.get)
	mux.HandleFunc("POST /api/v1/transfers", h.transfer)
}

type createReq struct {
	Owner        string          `json:"owner"`
	BalanceCents json.RawMessage `json:"balance_cents"`
}

func parseBalanceCents(raw json.RawMessage) (int64, string) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, ""
	}
	var num json.Number
	if err := json.Unmarshal(raw, &num); err != nil {
		return 0, "must be int64"
	}
	s := num.String()
	if strings.ContainsAny(s, ".eE") {
		return 0, "must be int64"
	}
	v, err := num.Int64()
	if err != nil {
		return 0, "must be int64"
	}
	return v, ""
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := decode(r, &req, h.maxBodyBytes); err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{Error: "invalid json"})
		return
	}
	req.Owner = strings.TrimSpace(req.Owner)
	balance, balanceErr := parseBalanceCents(req.BalanceCents)
	fields := map[string]string{}
	if req.Owner == "" {
		fields["owner"] = "required"
	}
	if balanceErr != "" {
		fields["balance_cents"] = balanceErr
	} else if balance < 0 {
		fields["balance_cents"] = "must be >= 0"
	}
	if len(fields) > 0 {
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{
			Error:  "validation failed",
			Fields: fields,
		})
		return
	}
	a, err := h.svc.Create(r.Context(), req.Owner, balance)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, httpapi.ErrorBody{Error: "create failed"})
		return
	}
	httpapi.WriteJSON(w, http.StatusCreated, a)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{Error: "invalid id"})
		return
	}
	a, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		httpapi.WriteJSON(w, http.StatusNotFound, httpapi.ErrorBody{Error: "not found"})
		return
	}
	if err != nil {
		httpapi.WriteJSON(w, http.StatusInternalServerError, httpapi.ErrorBody{Error: "get failed"})
		return
	}
	httpapi.WriteJSON(w, http.StatusOK, a)
}

func (h *Handler) transfer(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{
			Error: "Idempotency-Key header required",
		})
		return
	}
	var req TransferRequest
	if err := decode(r, &req, h.maxBodyBytes); err != nil {
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{Error: "invalid json"})
		return
	}
	res, err := h.svc.Transfer(r.Context(), key, req)
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteJSON(w, http.StatusNotFound, httpapi.ErrorBody{Error: "account not found"})
	case errors.Is(err, ErrInsufficientFunds):
		httpapi.WriteJSON(w, http.StatusConflict, httpapi.ErrorBody{Error: "insufficient funds"})
	case errors.Is(err, ErrConflict):
		httpapi.WriteJSON(w, http.StatusConflict, httpapi.ErrorBody{
			Error: "version conflict",
			Hint:  "reload accounts and retry",
		})
	case errors.Is(err, ErrIdempotencyConflict):
		httpapi.WriteJSON(w, http.StatusConflict, httpapi.ErrorBody{
			Error: "idempotency key reuse with different body",
		})
	case err != nil:
		httpapi.WriteJSON(w, http.StatusBadRequest, httpapi.ErrorBody{Error: err.Error()})
	default:
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if res.Replay {
			w.Header().Set("Idempotent-Replayed", "true")
		}
		w.WriteHeader(res.StatusCode)
		_, _ = w.Write(res.Body)
	}
}

func decode(r *http.Request, dst any, max int64) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, max))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
