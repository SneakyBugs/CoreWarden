package storage

import (
	"context"
	"database/sql"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sneakybugs/corewarden/api/database"
	"github.com/sneakybugs/corewarden/api/database/queries"
	"github.com/sneakybugs/corewarden/api/services/health"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	migrationInitialDelay = time.Second
	migrationMaxDelay     = 30 * time.Second
)

type Options struct {
	ConnectionString string
}

func NewService(lc fx.Lifecycle, options Options, rc *health.ReadinessChecks, logger *zap.Logger) (Storage, error) {
	pool, err := pgxpool.New(context.Background(), options.ConnectionString)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDBFromPool(pool)
	var migrated atomic.Bool
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					if migrateWithBackoff(ctx, logger, db) {
						migrated.Store(true)
					}
				}()
				return nil
			},
			OnStop: func(context.Context) error {
				cancel()
				pool.Close()
				return nil
			},
		},
	)
	rc.Add(&readinessCheck{
		db:       db,
		migrated: &migrated,
	})
	return &PostgresStorage{queries: queries.New(pool), pool: pool}, nil
}

func migrateWithBackoff(ctx context.Context, logger *zap.Logger, db *sql.DB) bool {
	migrations := database.GetMigrations()
	delay := migrationInitialDelay
	for attempt := 1; ; attempt++ {
		_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
		if err == nil {
			logger.Info("database migrated")
			return true
		}
		if ctx.Err() != nil {
			return false
		}
		logger.Warn(
			"database migration failed",
			zap.Int("attempt", attempt),
			zap.Duration("nextDelay", delay),
			zap.Error(err),
		)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return false
		case <-timer.C:
		}
		delay = min(delay*2, migrationMaxDelay)
	}
}

type readinessCheck struct {
	db       *sql.DB
	migrated *atomic.Bool
}

func (rc *readinessCheck) Ready() bool {
	return rc.migrated.Load() && rc.db.Ping() == nil
}
