package service

import (
	"context"
	"fmt"

	"go_sql_mid_trainer_v2/internal/domain"
)

type Repository interface {
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	SearchUsers(ctx context.Context, q string, limit int) ([]domain.User, error)
	ListOrdersCursor(ctx context.Context, userID int64, cursor *domain.OrderCursor, limit int) ([]domain.Order, *domain.OrderCursor, error)
	ListItemsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error)
	CreateTransferTx(ctx context.Context, req domain.TransferRequest, idempotencyKey string) (domain.Transfer, error)
}

type RiskClient interface {
	GetRiskProfile(ctx context.Context, userID int64) (domain.RiskProfile, error)
}

type Service struct {
	repo Repository
	risk RiskClient
}

func New(repo Repository, risk RiskClient) *Service {
	return &Service{repo: repo, risk: risk}
}

func (s *Service) GetUser(ctx context.Context, id int64) (domain.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) SearchUsers(ctx context.Context, q string, limit int) ([]domain.User, error) {
	return s.repo.SearchUsers(ctx, q, limit)
}

// TODO #8.
// Собери страницу заказов с items.
// Требования:
//   - вызови repo.ListOrdersCursor;
//   - если заказов нет, верни пустой page и nextCursor как есть;
//   - собери orderIDs;
//   - одним вызовом repo.ListItemsByOrderIDs получи items;
//   - собери []domain.OrderDTO в порядке orders;
//   - TotalCents = сумма item.Qty * item.PriceCents;
//   - ошибки оборачивай через fmt.Errorf с %w.
func (s *Service) GetOrdersWithItems(ctx context.Context, userID int64, cursor *domain.OrderCursor, limit int) (domain.OrderPage, error) {
	return domain.OrderPage{}, domain.ErrNotImplemented
}

// TODO #9.
// Валидируй и выполни перевод.
// Требования:
//   - from/to > 0;
//   - from != to;
//   - amount > 0;
//   - idempotencyKey не пустой;
//   - затем вызови repo.CreateTransferTx;
//   - domain.ErrInvalidInput/domain.ErrIdempotencyMissing для плохого входа;
//   - технические ошибки wrap через %w, доменные можно вернуть как есть или тоже wrap, но так, чтобы errors.Is работал.
func (s *Service) Transfer(ctx context.Context, req domain.TransferRequest, idempotencyKey string) (domain.Transfer, error) {
	return domain.Transfer{}, domain.ErrNotImplemented
}

// TODO #10.
// Получи пользователя и risk profile параллельно.
// Требования:
//   - создай child context с cancel;
//   - параллельно вызови repo.GetUserByID и risk.GetRiskProfile;
//   - при первой ошибке отменяй context;
//   - дождись обе goroutine;
//   - не допускай data race;
//   - если одна операция вернула ошибку, верни её с контекстом;
//   - если всё хорошо, верни domain.RiskSummary.
//
// Можно решить через channels + WaitGroup. errgroup специально не подключён, чтобы ты руками вспомнил concurrency, этот маленький аттракцион боли.
func (s *Service) RiskSummary(ctx context.Context, userID int64) (domain.RiskSummary, error) {
	return domain.RiskSummary{}, domain.ErrNotImplemented
}

var _ = fmt.Errorf
