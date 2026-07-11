package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go_sql_mid_trainer_v2/internal/domain"
)

type Handler func(ctx context.Context, job domain.EmailJob) error

type ErrorMode int

const (
	FailFast ErrorMode = iota
	CollectAll
)

// TODO #17.
// Реализуй worker pool.
// Требования:
//   - workers <= 0 -> workers = 1;
//   - запусти workers goroutine;
//   - каждая читает jobs из канала до закрытия или ctx.Done();
//   - handle вызывай с child context;
//   - если mode == FailFast: первая ошибка вызывает cancel и возвращается наружу;
//   - если mode == CollectAll: дождись всех, верни errors.Join(errs...);
//   - не пиши в общий slice ошибок без mutex;
//   - не допускай goroutine leak;
//   - если ctx отменён до ошибок, верни ctx.Err().
func RunPool(ctx context.Context, workers int, jobs <-chan domain.EmailJob, mode ErrorMode, handle Handler) error {
	return domain.ErrNotImplemented
}

var _ = sync.Mutex{}
var _ = errors.Join
var _ = fmt.Errorf
