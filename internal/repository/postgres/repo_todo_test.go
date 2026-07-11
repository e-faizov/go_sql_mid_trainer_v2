//go:build todo

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"go_sql_mid_trainer_v2/internal/domain"
	"go_sql_mid_trainer_v2/internal/repository/postgres"
)

func testRepo(t *testing.T) *postgres.Repo {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is empty")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	return postgres.New(db)
}

func TestGetUserByIDTODO(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	user, err := repo.GetUserByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserByID(1): %v", err)
	}
	if user.ID != 1 || user.Name != "Alice" || user.Email == "" {
		t.Fatalf("unexpected user: %+v", user)
	}

	_, err = repo.GetUserByID(ctx, 999)
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	_, err = repo.GetUserByID(ctx, 0)
	if !errors.Is(err, domain.ErrWrongID) {
		t.Fatalf("expected ErrWrongID, got %v", err)
	}
}

func TestListItemsByOrderIDsTODO(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	items, err := repo.ListItemsByOrderIDs(ctx, []int64{1, 2})
	if err != nil {
		t.Fatalf("ListItemsByOrderIDs: %v", err)
	}
	if len(items[1]) != 2 {
		t.Fatalf("order 1 should have 2 items, got %+v", items[1])
	}
	if len(items[2]) != 2 {
		t.Fatalf("order 2 should have 2 items, got %+v", items[2])
	}
	if _, ok := items[999]; ok {
		t.Fatalf("unexpected group for unknown order id")
	}

	empty, err := repo.ListItemsByOrderIDs(ctx, nil)
	if err != nil {
		t.Fatalf("empty ids should not error: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty ids should return empty map, got %+v", empty)
	}
}

func TestListOrdersCursorTODO(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()

	orders, next, err := repo.ListOrdersCursor(ctx, 1, nil, 2)
	if err != nil {
		t.Fatalf("ListOrdersCursor page1: %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("want 2 orders, got %d", len(orders))
	}
	if next == nil {
		t.Fatalf("want next cursor")
	}
	if !orders[0].CreatedAt.After(orders[1].CreatedAt) && !orders[0].CreatedAt.Equal(orders[1].CreatedAt) {
		t.Fatalf("orders are not sorted desc: %+v", orders)
	}

	orders2, _, err := repo.ListOrdersCursor(ctx, 1, next, 2)
	if err != nil {
		t.Fatalf("ListOrdersCursor page2: %v", err)
	}
	if len(orders2) == 0 {
		t.Fatalf("want second page")
	}
	if orders2[0].ID == orders[0].ID || orders2[0].ID == orders[1].ID {
		t.Fatalf("cursor did not advance, page2=%+v page1=%+v", orders2, orders)
	}
}

func TestCreateTransferTxTODO(t *testing.T) {
	repo := testRepo(t)
	ctx := context.Background()
	key := "test-key-" + time.Now().Format("150405.000000000")

	tr, err := repo.CreateTransferTx(ctx, domain.TransferRequest{
		FromAccountID: 1,
		ToAccountID:   2,
		AmountCents:   123,
	}, key)
	if err != nil {
		t.Fatalf("CreateTransferTx: %v", err)
	}
	if tr.ID == 0 || tr.IdempotencyKey != key || tr.AmountCents != 123 {
		t.Fatalf("bad transfer: %+v", tr)
	}

	tr2, err := repo.CreateTransferTx(ctx, domain.TransferRequest{
		FromAccountID: 1,
		ToAccountID:   2,
		AmountCents:   123,
	}, key)
	if err != nil {
		t.Fatalf("idempotent CreateTransferTx: %v", err)
	}
	if tr2.ID != tr.ID {
		t.Fatalf("idempotency should return existing transfer, got %+v vs %+v", tr2, tr)
	}
}
