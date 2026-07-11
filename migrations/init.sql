CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_created_id ON orders(user_id, created_at DESC, id DESC);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    name TEXT NOT NULL,
    qty INT NOT NULL CHECK (qty > 0),
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0)
);

CREATE INDEX idx_order_items_order_id ON order_items(order_id, id);

CREATE TABLE accounts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    balance_cents BIGINT NOT NULL CHECK (balance_cents >= 0)
);

CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    from_account_id BIGINT NOT NULL REFERENCES accounts(id),
    to_account_id BIGINT NOT NULL REFERENCES accounts(id),
    amount_cents BIGINT NOT NULL CHECK (amount_cents > 0),
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE email_jobs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO users(name, email, created_at) VALUES
('Alice', 'alice@example.com', now() - interval '5 days'),
('Bob', 'bob@example.com', now() - interval '4 days'),
('Carol', 'carol@example.com', now() - interval '3 days'),
('Dave', 'dave@example.com', now() - interval '2 days');

INSERT INTO orders(user_id, status, created_at) VALUES
(1, 'new', now() - interval '5 hours'),
(1, 'paid', now() - interval '4 hours'),
(1, 'shipped', now() - interval '3 hours'),
(2, 'new', now() - interval '2 hours'),
(3, 'paid', now() - interval '1 hour');

INSERT INTO order_items(order_id, name, qty, price_cents) VALUES
(1, 'book', 1, 1200),
(1, 'pen', 2, 150),
(2, 'phone', 1, 70000),
(2, 'case', 1, 2500),
(3, 'keyboard', 1, 9000),
(4, 'mouse', 1, 3500),
(5, 'monitor', 2, 30000);

INSERT INTO accounts(user_id, balance_cents) VALUES
(1, 100000),
(2, 25000),
(3, 5000),
(4, 0);

INSERT INTO email_jobs(user_id, kind) VALUES
(1, 'welcome'),
(2, 'promo'),
(3, 'receipt');
