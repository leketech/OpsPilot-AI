package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"

	// ⚠️ IMPORTANT: Update this import path to match where your config package actually lives!
	// If your config is in a folder named 'pkg/config', change it to:
	// "github.com/leketech/OpsPilot-AI/backend/pkg/config"
	"github.com/leketech/OpsPilot-AI/backend/internal/config"
)

// ConnectPostgres creates a connection pool to PostgreSQL
func ConnectPostgres(cfg *config.Config) (*pgxpool.Pool, error) {
	
	// 1. Build the connection string
	// Note: We add ?sslmode=disable because local Docker databases usually don't have SSL configured.
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// 2. Parse the config to allow custom pool settings
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	// Optional: Optimize the pool for your web server
	poolConfig.MaxConns = 10 // Maximum number of connections in the pool
	poolConfig.MinConns = 2  // Minimum number of idle connections

	// 3. Create the connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// 4. Test the connection to ensure the database is actually reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close() // Clean up the pool if the ping fails
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("✅ Successfully connected to PostgreSQL database pool!")
	return pool, nil
}