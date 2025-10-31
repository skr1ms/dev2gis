package storage

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skr1ms/dev2gis/config"
	"github.com/skr1ms/dev2gis/internal/storage/sqlc"
	"github.com/skr1ms/dev2gis/pkg/middleware"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
	logger  middleware.LoggerInterface
}

func NewDB(ctx context.Context, cfg config.Config, logger middleware.LoggerInterface) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName)

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 5 * time.Minute
	poolConfig.MaxConnIdleTime = 2 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		Pool:    pool,
		Queries: sqlc.New(pool),
		logger:  logger,
	}

	if err := db.RunMigrations(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database connected successfully", map[string]interface{}{
		"max_conns": poolConfig.MaxConns,
		"min_conns": poolConfig.MinConns,
	})

	return db, nil
}

func (db *DB) RunMigrations() error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	config := db.Pool.Config()
	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.ConnConfig.User, config.ConnConfig.Password,
		config.ConnConfig.Host, config.ConnConfig.Port,
		config.ConnConfig.Database)

	migrator, err := migrate.NewWithSourceInstance("iofs", sourceDriver, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	db.logger.Info("Database migrations completed successfully")
	return nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) GetPool() *pgxpool.Pool {
	return db.Pool
}

func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return db.Pool.Begin(ctx)
}

func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	var result int
	if err := db.Pool.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("database query health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database health check returned unexpected result: %d", result)
	}

	return nil
}

func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}

func (db *DB) LogStats() {
	stats := db.GetStats()
	db.logger.Debug("Database Pool Stats", map[string]interface{}{
		"total_conns":        stats.TotalConns(),
		"idle_conns":         stats.IdleConns(),
		"acquired_conns":     stats.AcquiredConns(),
		"constructing_conns": stats.ConstructingConns(),
	})
}

func (db *DB) GetQueries() *sqlc.Queries {
	return db.Queries
}

func (db *DB) WithTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := db.Queries.WithTx(tx)
	if err := fn(queries); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
