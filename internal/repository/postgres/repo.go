package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go_sql_mid_trainer_v2/internal/domain"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) DB() *sql.DB {
	return r.db
}

// TODO #1.
// Реализуй чтение пользователя по id.
// Требования:
//   - если id <= 0, верни domain.ErrWrongID;
//   - используй QueryRowContext;
//   - SELECT id, name, email, created_at FROM users WHERE id = $1;
//   - sql.ErrNoRows преобразуй в domain.ErrUserNotFound;
//   - остальные ошибки оборачивай через fmt.Errorf с %w.
func (r *Repo) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	if id <= 0 {
		return domain.User{}, domain.ErrWrongID
	}

	var res domain.User

	query := "SELECT id, name, email, created_at FROM users WHERE id = $1;"
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&res.ID, &res.Name, &res.Email, &res.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("query id: %d, %w", id, err)
	}

	return res, nil
}

// TODO #2.
// Реализуй поиск пользователей.
// Требования:
//   - если limit <= 0 или limit > 100, поставь 20;
//   - если q пустой, верни последних пользователей:
//     SELECT id, name, email, created_at FROM users ORDER BY created_at DESC, id DESC LIMIT $1;
//   - если q не пустой, ищи по name/email через ILIKE:
//     WHERE name ILIKE $1 OR email ILIKE $1;
//   - используй QueryContext;
//   - defer rows.Close();
//   - rows.Scan внутри цикла;
//   - rows.Err после цикла;
//   - пустой результат это не ошибка.
func (r *Repo) SearchUsers(ctx context.Context, q string, limit int) ([]domain.User, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	q = strings.TrimSpace(q)

	var rows *sql.Rows
	var err error

	if len(q) == 0 {
		query :=
			`SELECT id, name, email, created_at
		FROM users
		ORDER BY created_at DESC, id DESC LIMIT $1;`
		rows, err = r.db.QueryContext(ctx, query, limit)
	} else {
		query :=
			`SELECT id, name, email, created_at
		FROM users
		WHERE name ILIKE $1 OR email ILIKE $1
		ORDER BY created_at DESC, id DESC LIMIT $2;`
		pattern := "%" + q + "%"
		rows, err = r.db.QueryContext(ctx, query, pattern, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query q: '%s', %w", q, err)
	}
	defer rows.Close()

	res := []domain.User{}
	for rows.Next() {
		tmp := domain.User{}

		err = rows.Scan(&tmp.ID, &tmp.Name, &tmp.Email, &tmp.CreatedAt)
		if err != nil {
			return res, fmt.Errorf("rows scan %w", err)
		}
		res = append(res, tmp)
	}

	if err = rows.Err(); err != nil {
		return res, fmt.Errorf("rows err %w", err)
	}

	return res, nil
}

// TODO #3.
// Реализуй keyset pagination заказов пользователя.
// Требования:
//   - если userID <= 0, верни domain.ErrWrongID;
//   - если limit <= 0 или limit > 100, поставь 20;
//   - доставай limit+1 строку, чтобы понять, есть ли следующая страница;
//   - сортировка ORDER BY created_at DESC, id DESC;
//   - если cursor == nil:
//     SELECT id, user_id, status, created_at
//     FROM orders
//     WHERE user_id = $1
//     ORDER BY created_at DESC, id DESC
//     LIMIT $2
//   - если cursor != nil, добавь условие:
//     AND (created_at, id) < ($2, $3)
//     и LIMIT станет $4;
//   - верни orders максимум limit штук;
//   - если была лишняя строка, верни nextCursor по последнему возвращённому заказу;
//   - QueryContext, rows.Close, rows.Err.
func (r *Repo) ListOrdersCursor(ctx context.Context, userID int64, cursor *domain.OrderCursor, limit int) ([]domain.Order, *domain.OrderCursor, error) {
	if userID <= 0 {
		return nil, nil, domain.ErrWrongID
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var rows *sql.Rows
	var err error

	if cursor == nil {
		query := `SELECT id, user_id, status, created_at
     FROM orders
     WHERE user_id = $1
     ORDER BY created_at DESC, id DESC
     LIMIT $2`
		rows, err = r.db.QueryContext(ctx, query, userID, limit+1)
	} else {
		query := `SELECT id, user_id, status, created_at
     FROM orders
     WHERE user_id = $1
	 AND (created_at, id) < ($2, $3)
     ORDER BY created_at DESC, id DESC
     LIMIT $4`
		rows, err = r.db.QueryContext(ctx, query, userID, cursor.CreatedAt, cursor.ID, limit+1)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("query id: %d, %w", userID, err)
	}

	defer rows.Close()

	var limitCount int
	var newCursor *domain.OrderCursor
	res := make([]domain.Order, 0, limit)
	for rows.Next() {
		var tmp domain.Order
		err = rows.Scan(&tmp.ID, &tmp.UserID, &tmp.Status, &tmp.CreatedAt)
		if err != nil {
			return nil, nil, fmt.Errorf("rows scan id: %d, %w", userID, err)
		}
		limitCount++
		if limitCount <= limit {
			res = append(res, tmp)
		} else {
			newCursor = &domain.OrderCursor{}
			newCursor.ID = res[len(res)-1].ID
			newCursor.CreatedAt = res[len(res)-1].CreatedAt
		}
	}
	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows err id: %d, %w", userID, err)
	}

	return res, newCursor, nil
}

// TODO #4.
// Реализуй получение items сразу для нескольких заказов.
// Требования:
//   - если orderIDs пустой, верни пустую map без запроса в БД;
//   - QueryContext;
//   - для pgx stdlib можно передать []int64 прямо в ANY($1):
//     SELECT id, order_id, name, qty, price_cents
//     FROM order_items
//     WHERE order_id = ANY($1)
//     ORDER BY order_id, id;
//   - сгруппируй в map[int64][]domain.OrderItem по OrderID;
//   - rows.Close и rows.Err.
func (r *Repo) ListItemsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error) {
	res := map[int64][]domain.OrderItem{}
	if len(orderIDs) == 0 {
		return res, nil
	}

	query := `SELECT id, order_id, name, qty, price_cents
     FROM order_items
     WHERE order_id = ANY($1)
     ORDER BY order_id, id;`

	rows, err := r.db.QueryContext(ctx, query, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("query %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tmp domain.OrderItem
		err = rows.Scan(&tmp.ID, &tmp.OrderID, &tmp.Name, &tmp.Qty, &tmp.PriceCents)
		if err != nil {
			return nil, fmt.Errorf("scan %w", err)
		}
		items := res[tmp.OrderID]
		items = append(items, tmp)
		res[tmp.OrderID] = items
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows scan %w", err)
	}

	return res, nil
}

// TODO #5.
// Реализуй денежный перевод в транзакции.
// Требования:
//   - req уже валидируется в service, но repo всё равно не должен слепо верить миру;
//   - BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted});
//   - defer rollback, commit только в конце;
//   - если transfer с таким idempotency_key уже есть, верни его без повторного списания;
//   - заблокируй два account через SELECT ... FOR UPDATE;
//   - блокируй в стабильном порядке по account id, чтобы снижать шанс deadlock;
//   - если денег недостаточно, верни domain.ErrInsufficientFunds;
//   - обнови balances;
//   - вставь transfer и верни созданную запись;
//   - все технические ошибки wrap через %w;
//   - sql.ErrNoRows для account можно преобразовать в domain.ErrWrongID.
func (r *Repo) CreateTransferTx(ctx context.Context, req domain.TransferRequest, idempotencyKey string) (domain.Transfer, error) {
	return domain.Transfer{}, domain.ErrNotImplemented
}

// TODO #6.
// Реализуй выборку queued email_jobs с блокировкой.
// Требования:
//   - BeginTx;
//   - SELECT id, user_id, kind FROM email_jobs
//     WHERE status = 'queued'
//     ORDER BY id
//     LIMIT $1
//     FOR UPDATE SKIP LOCKED;
//   - поменяй status на 'processing', attempts = attempts + 1, updated_at = now();
//   - commit;
//   - rows.Close, rows.Err;
//   - пустой список это не ошибка.
func (r *Repo) LeaseEmailJobs(ctx context.Context, limit int) ([]domain.EmailJob, error) {
	return nil, domain.ErrNotImplemented
}

// TODO #7.
// Реализуй отметку email job как done/failed.
// Требования:
//   - status разрешён только done или queued;
//   - done при успехе;
//   - queued при временной ошибке, чтобы повторить позже;
//   - ExecContext;
//   - проверь RowsAffected: если 0, верни domain.ErrOrderNotFound или domain.ErrWrongID.
func (r *Repo) FinishEmailJob(ctx context.Context, jobID int64, status string) error {
	return domain.ErrNotImplemented
}

func normalizeLimit(limit int) int {
	if limit <= 0 || limit > 100 {
		return 20
	}
	return limit
}

// ensure imports stay useful while TODOs are stubs. Уберёшь это, когда реализуешь методы.
var _ = fmt.Errorf
var _ = sql.ErrNoRows
var _ = time.Time{}
