package storage

import (
	"context"
	"embed"
	"fmt"
	"snipz/internal/utils"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*sql
var migrationsFS embed.FS

type DB struct {
	*pgxpool.Pool
	QueryBuilder *squirrel.StatementBuilderType
	url          string
}

func New(ctx context.Context, config *utils.DB) (*DB, error) {
	url := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable",
		config.Connection, config.User, config.Password, config.Host, config.Port, config.Name,
	)

	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, err
	}

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &DB{db, &psql, url}, nil
}

func (db *DB) Migrate() error {
	driver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	migrations, err := migrate.NewWithSourceInstance("iofs", driver, db.url)
	if err != nil {
		return err
	}

	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func (db *DB) ErrorCode(err error) string {
	pgErr := err.(*pgconn.PgError)
	return pgErr.Code
}

func (db *DB) Close() {
	db.Pool.Close()
}
