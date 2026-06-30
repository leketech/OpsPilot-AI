package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	// ⚠️ IMPORTANT: Update this import path to match where your config package actually lives!
	"github.com/leketech/OpsPilot-AI/backend/internal/config"
)

// ConnectRedis creates a Redis client and verifies the connection
func ConnectRedis(cfg *config.Config) (*redis.Client, error) {

	// 1. Create the Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		// Password: cfg.RedisPassword, // Uncomment and add to config if your Redis requires a password
		DB:       0, // Use default database (0-15)
	})

	// 2. Test the connection to ensure Redis is actually reachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info("✅ Successfully connected to Redis!")
	return rdb, nil
}