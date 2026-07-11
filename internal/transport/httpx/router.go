package httpx

import (
	"net/http"

	"go_sql_mid_trainer_v2/internal/service"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /users/search", h.handleSearchUsers)
	mux.HandleFunc("GET /users/{id}", h.handleGetUser)
	mux.HandleFunc("GET /users/{id}/orders", h.handleListOrders)
	mux.HandleFunc("GET /users/{id}/risk", h.handleRiskSummary)
	mux.HandleFunc("POST /transfers", h.handleCreateTransfer)

	// Mock внешнего сервиса. Реальный ExternalClient ходит сюда же по HTTP.
	mux.HandleFunc("GET /external/risk/{id}", h.handleExternalRiskMock)

	return loggingMiddleware(mux)
}
