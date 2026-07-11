package worker

import (
	"context"
	"errors"
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
	if workers <= 0 {
		workers = 1
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var resError error
	chErr := make(chan error, workers)

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case emailJob, ok := <-jobs:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					default:
					}

					err := handle(ctx, emailJob)
					if err != nil {
						chErr <- err
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(chErr)
	}()

	allErrors := []error{}

loop:
	for {
		err, ok := <-chErr
		if ok {
			if mode == FailFast {
				if err != nil && resError == nil {
					resError = err
					cancel()
				}
			} else {
				allErrors = append(allErrors, err)
			}
		} else {
			break loop
		}
	}

	if len(allErrors) != 0 {
		resError = errors.Join(allErrors...)
	} else if err := ctx.Err(); resError == nil && err != nil {
		resError = err
	}

	return resError
}
