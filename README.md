# Go + SQL Middle Trainer v2

Мини-проект для тренировки Go middle: `database/sql`, PostgreSQL, HTTP handlers, внешний HTTP-клиент, `context`, транзакции, каналы и worker pool.

Это не решение. Это почти готовая обвязка с TODO-местами. Да, придётся писать код руками. Трагедия века, но зато навык появится.

## Запуск

```bash
docker compose up -d
go mod tidy
go run ./cmd/api
```

Сервер стартует на `:8081`, Postgres проброшен на `localhost:55433`.

Проверка:

```bash
curl http://localhost:8081/health
curl http://localhost:8081/users/1
curl 'http://localhost:8081/users/1/orders?limit=2'
curl 'http://localhost:8081/users/search?q=a&limit=10'
curl http://localhost:8081/users/1/risk
curl -X POST http://localhost:8081/transfers \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-key-1' \
  -d '{"from_account_id":1,"to_account_id":2,"amount_cents":500}'
```

Пока TODO не реализованы, часть endpoint'ов вернёт `501`.

## Что нужно реализовать

### Repository, `internal/repository/postgres/repo.go`

1. `GetUserByID`
   - `QueryRowContext`
   - `sql.ErrNoRows -> domain.ErrUserNotFound`
   - wrapping через `%w`

2. `SearchUsers`
   - `QueryContext`
   - фильтр по `ILIKE`, если `q` не пустой
   - `LIMIT`, защита от плохого limit
   - `rows.Close()` и `rows.Err()`

3. `ListOrdersCursor`
   - keyset pagination по `(created_at, id)`
   - сортировка `created_at DESC, id DESC`
   - вернуть `nextCursor`, если записей больше limit

4. `ListItemsByOrderIDs`
   - пустой список без запроса в БД
   - `WHERE order_id = ANY($1)`
   - группировка в `map[int64][]domain.OrderItem`

5. `CreateTransferTx`
   - транзакция через `BeginTx`
   - idempotency key
   - `SELECT ... FOR UPDATE`
   - порядок блокировок по account id
   - проверка баланса
   - update балансов
   - insert transfer
   - commit/rollback

### Service, `internal/service/service.go`

6. `GetOrdersWithItems`
   - взять страницу заказов
   - одним запросом взять items по order IDs
   - собрать DTO
   - посчитать `TotalCents`

7. `Transfer`
   - валидация входа
   - `from != to`
   - `amount > 0`
   - idempotency key обязателен
   - вызвать repo transaction

8. `RiskSummary`
   - параллельно получить пользователя из repo и risk profile из external client
   - корректно обработать ошибки
   - отменять оставшуюся работу при первой ошибке

### HTTP transport, `internal/transport/httpx/handlers.go`

9. `handleGetUser`
   - парсинг path id
   - mapping ошибок в HTTP status
   - безопасный JSON error response

10. `handleListOrders`
    - path id
    - query params `limit`, `cursor_created_at`, `cursor_id`
    - JSON response

11. `handleCreateTransfer`
    - decode JSON
    - header `Idempotency-Key`
    - status codes: 201/200/400/409/500

### External client, `internal/client/external/client.go`

12. `GetRiskProfile`
    - `http.NewRequestWithContext`
    - `client.Do`
    - `defer resp.Body.Close()`
    - обработка статусов
    - decode JSON

### Worker, `internal/worker/pool.go`

13. `RunPool`
    - ограниченное число worker'ов
    - слушать `ctx.Done()`
    - читать jobs до закрытия канала
    - собрать первую ошибку или все ошибки, по выбранной стратегии
    - не допустить goroutine leak

## Тесты

Обычные тесты:

```bash
go test ./...
```

TODO/integration тесты включаются тегом:

```bash
docker compose up -d
TEST_DATABASE_URL='postgres://app:app@localhost:55433/app?sslmode=disable' go test -tags todo ./...
```

Они специально будут падать, пока ты не реализуешь TODO. Вот такая подлость, но полезная.

## Подсказка по DSN

Если приложение пытается подключиться к `localhost:5432`, значит у тебя выставлен `DATABASE_URL`. Проверь:

```bash
echo $DATABASE_URL
```

Сбросить:

```bash
unset DATABASE_URL
```
