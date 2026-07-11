package service

import (
	"context"
	"fmt"

	"go_sql_mid_trainer_v2/internal/domain"
)

type Repository interface {
	GetUserByID(ctx context.Context, id int64) (domain.User, error)
	SearchUsers(ctx context.Context, q string, limit int) ([]domain.User, error)
	ListOrdersCursor(ctx context.Context,
		userID int64, cursor *domain.OrderCursor, limit int) ([]domain.Order, *domain.OrderCursor, error)
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
func (s *Service) GetOrdersWithItems(ctx context.Context,
	userID int64, cursor *domain.OrderCursor, limit int) (domain.OrderPage, error) {
	orders, nextCursor, err := s.repo.ListOrdersCursor(ctx, userID, cursor, limit)
	if err != nil {
		return domain.OrderPage{}, fmt.Errorf("GetOrdersWithItems id: %d, %w", userID, err)
	}

	if len(orders) == 0 {
		return domain.OrderPage{
			NextCursor: nextCursor,
		}, nil
	}

	orderIDs := make([]int64, 0, len(orders))
	for _, v := range orders {
		orderIDs = append(orderIDs, v.ID)
	}

	mapItems, err := s.repo.ListItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return domain.OrderPage{}, fmt.Errorf("ListItemsByOrderIDs id: %d, %w", userID, err)
	}

	orderDTO := make([]domain.OrderDTO, 0, len(orders))
	for _, v := range orders {
		items := mapItems[v.ID]
		var totalCents int64
		for _, i := range items {
			totalCents += int64(i.Qty) * i.PriceCents
		}
		tmp := domain.OrderDTO{
			ID:         v.ID,
			Status:     v.Status,
			CreatedAt:  v.CreatedAt,
			Items:      items,
			TotalCents: totalCents,
		}
		orderDTO = append(orderDTO, tmp)

	}

	return domain.OrderPage{
		Orders:     orderDTO,
		NextCursor: nextCursor,
	}, nil
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
func (s *Service) Transfer(ctx context.Context,
	req domain.TransferRequest, idempotencyKey string) (domain.Transfer, error) {
	if req.FromAccountID <= 0 ||
		req.ToAccountID <= 0 ||
		req.AmountCents <= 0 ||
		req.FromAccountID == req.ToAccountID {
		return domain.Transfer{}, domain.ErrInvalidInput
	}
	if len(idempotencyKey) == 0 {
		return domain.Transfer{}, domain.ErrIdempotencyMissing
	}

	transfer, err := s.repo.CreateTransferTx(ctx, req, idempotencyKey)
	if err != nil {
		return domain.Transfer{}, fmt.Errorf("create transfer from: %d, to: %d, %w", req.FromAccountID, req.ToAccountID, err)
	}

	return transfer, nil
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
// Можно решить через channels + WaitGroup. errgroup специально не подключён,
//
//	чтобы ты руками вспомнил concurrency, этот маленький аттракцион боли.
func (s *Service) RiskSummary(ctx context.Context, userID int64) (domain.RiskSummary, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type withError[T any] struct {
		data T
		err  error
	}

	//var wg sync.WaitGroup

	chUser := make(chan withError[domain.User])
	chRisk := make(chan withError[domain.RiskProfile])

	//wg.Add(1)
	go func() {
		//defer wg.Done()
		user, err := s.repo.GetUserByID(ctx, userID)
		if err != nil {
			err = fmt.Errorf("get user by id: %w", err)
		}
		chUser <- withError[domain.User]{
			data: user,
			err:  err,
		}
	}()

	//wg.Add(1)
	go func() {
		//defer wg.Done()
		risk, err := s.risk.GetRiskProfile(ctx, userID)
		if err != nil {
			err = fmt.Errorf("get risky profile: %w", err)
		}
		chRisk <- withError[domain.RiskProfile]{
			data: risk,
			err:  err,
		}
	}()

	res := domain.RiskSummary{}
	var err error

	for chRisk != nil || chUser != nil {
		select {
		case data, ok := <-chUser:
			chUser = nil
			if ok {
				if data.err != nil {
					cancel()
					if err == nil {
						err = data.err
					}
					continue
				}
				res.User = data.data
			}
		case data, ok := <-chRisk:
			chRisk = nil
			if ok {
				if data.err != nil {
					cancel()
					if err == nil {
						err = data.err
					}
					continue
				}
				res.Risk = data.data
			}
		}
	}

	//wg.Wait()
	if err != nil {
		return domain.RiskSummary{}, err
	}

	return res, nil
}
