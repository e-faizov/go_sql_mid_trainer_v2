package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"go_sql_mid_trainer_v2/internal/client/external"
	"go_sql_mid_trainer_v2/internal/config"
	"go_sql_mid_trainer_v2/internal/domain"
	"go_sql_mid_trainer_v2/internal/repository/postgres"
	"go_sql_mid_trainer_v2/internal/service"
	"go_sql_mid_trainer_v2/internal/transport/httpx"
	"go_sql_mid_trainer_v2/internal/worker"
)

func main() {
	cfg := config.Load()

	db, err := openDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := postgres.New(db)
	riskClient := external.New(cfg.ExternalBaseURL, &http.Client{Timeout: 3 * time.Second})
	svc := service.New(repo, riskClient)
	h := httpx.NewHandler(svc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startDemoWorker(ctx)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("http listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(10 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func startDemoWorker(ctx context.Context) {
	jobs := make(chan domain.EmailJob)

	go func() {
		defer close(jobs)
		for i := int64(1); i <= 3; i++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- domain.EmailJob{ID: i, UserID: i, Kind: "demo"}:
			}
		}
	}()

	go func() {
		err := worker.RunPool(ctx, 2, jobs, worker.CollectAll, func(ctx context.Context, job domain.EmailJob) error {
			log.Printf("demo worker would process job id=%d kind=%s", job.ID, job.Kind)
			return nil
		})
		if err != nil && !errors.Is(err, domain.ErrNotImplemented) && !errors.Is(err, context.Canceled) {
			log.Printf("demo worker: %v", err)
		}
	}()
}
