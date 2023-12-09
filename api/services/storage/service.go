package storage

import (
	"context"

	"git.houseofkummer.com/lior/home-dns/api/database"
	"git.houseofkummer.com/lior/home-dns/api/database/queries"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/fx"
)

type Options struct {
	ConnectionString string
}

func NewService(lc fx.Lifecycle, options Options) (Storage, error) {
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
	return &PostgresStorage{queries: queries.New(pool)}, nil
}
