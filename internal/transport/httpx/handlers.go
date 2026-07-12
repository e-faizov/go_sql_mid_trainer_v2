package httpx

import (
	"encoding/json"
	"net/http"

	"go_sql_mid_trainer_v2/internal/domain"
)

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// TODO #12.
// Реализуй handler поиска пользователей.
// Требования:
//   - query params: q, limit;
//   - вызови h.svc.SearchUsers(r.Context(), q, limit);
//   - ошибки мапь через statusFromError/safeMessage;
//   - успешный ответ JSON: {"users": [...]}.
func (h *Handler) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query().Get("q")
	limit := parseLimit(r)

	users, err := h.svc.SearchUsers(ctx, q, limit)
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	if users == nil {
		users = []domain.User{}
	}

	res := struct {
		Users []domain.User `json:"users"`
	}{
		Users: users,
	}

	writeJSON(w, http.StatusOK, res)
}

// TODO #13.
// Реализуй GetUser handler.
// Требования:
//   - id взять из r.PathValue("id");
//   - провалидировать id > 0;
//   - вызвать service;
//   - не отдавать err.Error() клиенту для 500;
//   - статус брать через statusFromError;
//   - успешный ответ JSON с user.
func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	we := func(err error) {
		writeError(w, statusFromError(err), safeMessage(err))
	}

	id, err := parsePositiveInt64(r.PathValue("id"))
	if err != nil {
		we(err)
		return
	}

	user, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		we(err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// TODO #14.
// Реализуй список заказов.
// Требования:
//   - path id -> userID;
//   - query limit;
//   - query cursor_created_at и cursor_id опциональны;
//   - parseCursor уже готов;
//   - вызвать h.svc.GetOrdersWithItems;
//   - успешный ответ JSON domain.OrderPage.
func (h *Handler) handleListOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePositiveInt64(r.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	limit := parseLimit(r)
	cursor, err := parseCursor(r)
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	page, err := h.svc.GetOrdersWithItems(r.Context(), userID, cursor, limit)
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, page)
}

// TODO #15.
// Реализуй создание перевода.
// Требования:
//   - decode JSON body в domain.TransferRequest;
//   - если JSON битый, вернуть 400;
//   - взять Idempotency-Key из header;
//   - вызвать h.svc.Transfer;
//   - domain.ErrInsufficientFunds -> 409;
//   - успешный статус 201;
//   - если повтор idempotency вернул уже существующий transfer, 200 тоже допустим, но для простоты можно 201.
func (h *Handler) handleCreateTransfer(w http.ResponseWriter, r *http.Request) {

	transferRequest := domain.TransferRequest{}
	err := json.NewDecoder(r.Body).Decode(&transferRequest)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	idempotencyKey := r.Header.Get("Idempotency-Key")

	transfer, err := h.svc.Transfer(r.Context(), transferRequest, idempotencyKey)
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	writeJSON(w, http.StatusCreated, transfer)
}

// TODO #16.
// Реализуй RiskSummary handler.
// Требования:
//   - path id;
//   - вызвать h.svc.RiskSummary;
//   - ошибки мапить безопасно;
//   - успешный JSON.
func (h *Handler) handleRiskSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePositiveInt64(r.PathValue("id"))
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	risk, err := h.svc.RiskSummary(r.Context(), userID)
	if err != nil {
		writeError(w, statusFromError(err), safeMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, risk)
}

func (h *Handler) handleExternalRiskMock(w http.ResponseWriter, r *http.Request) {
	id, err := parsePositiveInt64(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "wrong id")
		return
	}

	if id == 13 {
		writeError(w, http.StatusServiceUnavailable, "risk service temporarily unavailable")
		return
	}

	level := "low"
	score := 10
	if id%2 == 0 {
		level = "medium"
		score = 45
	}
	if id%5 == 0 {
		level = "high"
		score = 80
	}

	writeJSON(w, http.StatusOK, domain.RiskProfile{
		UserID: id,
		Level:  level,
		Score:  score,
	})
}
