//go:build todo

package worker_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"go_sql_mid_trainer_v2/internal/domain"
	"go_sql_mid_trainer_v2/internal/worker"
)

func TestRunPoolCollectAllTODO(t *testing.T) {
	jobs := make(chan domain.EmailJob)
	go func() {
		defer close(jobs)
		for i := int64(1); i <= 5; i++ {
			jobs <- domain.EmailJob{ID: i, UserID: i, Kind: "test"}
		}
	}()

	var handled atomic.Int64
	err := worker.RunPool(context.Background(), 3, jobs, worker.CollectAll, func(ctx context.Context, job domain.EmailJob) error {
		handled.Add(1)
		if job.ID%2 == 0 {
			return fmt.Errorf("job %d failed", job.ID)
		}
		return nil
	})
	if err == nil {
		t.Fatalf("expected joined errors")
	}
	if handled.Load() != 5 {
		t.Fatalf("collect-all should handle all jobs, handled=%d", handled.Load())
	}
}

func TestRunPoolFailFastTODO(t *testing.T) {
	jobs := make(chan domain.EmailJob)
	go func() {
		defer close(jobs)
		for i := int64(1); i <= 100; i++ {
			jobs <- domain.EmailJob{ID: i, UserID: i, Kind: "test"}
		}
	}()

	sentinel := errors.New("boom")
	err := worker.RunPool(context.Background(), 4, jobs, worker.FailFast, func(ctx context.Context, job domain.EmailJob) error {
		if job.ID == 3 {
			return sentinel
		}
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}
}
