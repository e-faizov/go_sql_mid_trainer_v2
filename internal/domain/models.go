package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Order struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderItem struct {
	ID         int64  `json:"id"`
	OrderID    int64  `json:"order_id"`
	Name       string `json:"name"`
	Qty        int    `json:"qty"`
	PriceCents int64  `json:"price_cents"`
}

type OrderDTO struct {
	ID         int64       `json:"id"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items"`
	TotalCents int64       `json:"total_cents"`
}

type OrderCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int64     `json:"id"`
}

type OrderPage struct {
	Orders     []OrderDTO   `json:"orders"`
	NextCursor *OrderCursor `json:"next_cursor,omitempty"`
}

type Account struct {
	ID           int64 `json:"id"`
	UserID       int64 `json:"user_id"`
	BalanceCents int64 `json:"balance_cents"`
}

type TransferRequest struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	AmountCents   int64 `json:"amount_cents"`
}

type Transfer struct {
	ID             int64     `json:"id"`
	FromAccountID  int64     `json:"from_account_id"`
	ToAccountID    int64     `json:"to_account_id"`
	AmountCents    int64     `json:"amount_cents"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreatedAt      time.Time `json:"created_at"`
}

type RiskProfile struct {
	UserID int64  `json:"user_id"`
	Level  string `json:"level"`
	Score  int    `json:"score"`
}

type RiskSummary struct {
	User User        `json:"user"`
	Risk RiskProfile `json:"risk"`
}

type EmailJob struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Kind   string `json:"kind"`
}
