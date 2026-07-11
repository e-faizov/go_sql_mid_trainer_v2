package httpx

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_sql_mid_trainer_v2/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func statusFromError(err error) int {
	switch {
	case errors.Is(err, domain.ErrWrongID), errors.Is(err, domain.ErrInvalidInput), errors.Is(err, domain.ErrIdempotencyMissing):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrOrderNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict), errors.Is(err, domain.ErrInsufficientFunds):
		return http.StatusConflict
	case errors.Is(err, domain.ErrNotImplemented):
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

func safeMessage(err error) string {
	if errors.Is(err, domain.ErrNotImplemented) {
		return "not implemented"
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		return "user not found"
	}
	if errors.Is(err, domain.ErrWrongID) {
		return "wrong id"
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		return "invalid input"
	}
	if errors.Is(err, domain.ErrIdempotencyMissing) {
		return "idempotency key is required"
	}
	if errors.Is(err, domain.ErrInsufficientFunds) {
		return "insufficient funds"
	}
	return "internal server error"
}

func parsePositiveInt64(s string) (int64, error) {
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		return 0, domain.ErrWrongID
	}
	return id, nil
}

func parseLimit(r *http.Request) int {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		return 20
	}
	return limit
}

func parseCursor(r *http.Request) (*domain.OrderCursor, error) {
	createdRaw := r.URL.Query().Get("cursor_created_at")
	idRaw := r.URL.Query().Get("cursor_id")
	if createdRaw == "" && idRaw == "" {
		return nil, nil
	}
	if createdRaw == "" || idRaw == "" {
		return nil, domain.ErrInvalidInput
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdRaw)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	id, err := parsePositiveInt64(idRaw)
	if err != nil {
		return nil, err
	}
	return &domain.OrderCursor{CreatedAt: createdAt, ID: id}, nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
	})
}
