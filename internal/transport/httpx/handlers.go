package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	writeError(w, http.StatusNotImplemented, "not implemented")
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
	writeError(w, http.StatusNotImplemented, "not implemented")
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
	writeError(w, http.StatusNotImplemented, "not implemented")
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
	writeError(w, http.StatusNotImplemented, "not implemented")
}

// TODO #16.
// Реализуй RiskSummary handler.
// Требования:
//   - path id;
//   - вызвать h.svc.RiskSummary;
//   - ошибки мапить безопасно;
//   - успешный JSON.
func (h *Handler) handleRiskSummary(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
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

// Эти var нужны только чтобы файл компилировался, пока handlers ещё TODO.
// Уберёшь, когда реализуешь функции выше. Да, костыль, но честный тренировочный.
var _ = json.NewDecoder
var _ = errors.Is
var _ = fmt.Errorf
var _ = strconv.ParseInt
var _ = domain.TransferRequest{}
