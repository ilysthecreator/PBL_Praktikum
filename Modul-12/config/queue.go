package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
)

// RedisOpt membaca alamat Redis dari .env. Objek yang sama dipakai oleh
// client (relay), server (worker), dan inspector (endpoint admin).
func RedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     GetEnv("REDIS_ADDR", "localhost:6379"),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvInt("REDIS_DB", 0),
		// Timeout dibuat pendek. Saat Redis mati, kita ingin segera
		// tahu, bukan menunggu lama dengan koneksi database yang ter-lock.
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

// NewWorkerServer menyiapkan server Asynq — "Fiber"-nya worker.
func NewWorkerServer(logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(RedisOpt(), asynq.Config{
		// Jumlah task yang dikerjakan bersamaan.
		Concurrency: GetEnvInt("WORKER_CONCURRENCY", 5),
		// Worker HANYA mengambil task dari queue yang disebut di sini.
		// PERBAIKAN BUG 8: Gunakan model.QueueNotifications ("notifications")
		Queues: map[string]int{model.QueueNotifications: 1},
		// Delay sebelum melakukan retry.
		RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
			return time.Duration((n+1)*2) * time.Second
		},
		ShutdownTimeout: 10 * time.Second,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, t *asynq.Task, err error) {
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			logger.Warn("task_failed",
				slog.String("type", t.Type()),
				slog.String("attempt", fmt.Sprintf("%d/%d", retry+1, maxRetry+1)),
				slog.String("error", err.Error()))
		}),
		Logger:   newAsynqLogger(logger),
		LogLevel: asynq.WarnLevel,
	})
}

// asynqLogger menyalurkan log internal Asynq ke slog.
type asynqLogger struct{ l *slog.Logger }

func newAsynqLogger(l *slog.Logger) asynqLogger {
	return asynqLogger{l: l.With(slog.String("component", "asynq"))}
}

func (a asynqLogger) Debug(args ...any) { a.l.Debug(fmt.Sprint(args...)) }
func (a asynqLogger) Info(args ...any)  { a.l.Info(fmt.Sprint(args...)) }
func (a asynqLogger) Warn(args ...any)  { a.l.Warn(fmt.Sprint(args...)) }
func (a asynqLogger) Error(args ...any) { a.l.Error(fmt.Sprint(args...)) }
func (a asynqLogger) Fatal(args ...any) {
	a.l.Error(fmt.Sprint(args...))
	os.Exit(1)
}

