package storage

import (
	"context"
	"database/sql"

	"git.houseofkummer.com/lior/home-dns/api/database"
	"git.houseofkummer.com/lior/home-dns/api/database/queries"
	"git.houseofkummer.com/lior/home-dns/api/services/health"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/fx"
)

type Options struct {
	ConnectionString string
}

func NewService(lc fx.Lifecycle, options Options, rc *health.ReadinessChecks) (Storage, error) {
	pool, err := pgxpool.New(context.Background(), options.ConnectionString)
	if err != nil {
		return nil, err
	}
	migrations := database.GetMigrations()
	db := stdlib.OpenDBFromPool(pool)
	_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		panic(err)
	}
	lc.Append(
		fx.Hook{OnStart: func(ctx context.Context) error {
			migrations := database.GetMigrations()
			db := stdlib.OpenDBFromPool(pool)
			_, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
			return err
		}},
	)
	lc.Append(
		fx.Hook{OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		}},
	)
	rc.Add(&readinessCheck{
		db: db,
	})
	return &PostgresStorage{queries: queries.New(pool), pool: pool}, nil
}

type readinessCheck struct {
	db *sql.DB
}

func (rc *readinessCheck) Ready() bool {
	return rc.db.Ping() == nil
}
