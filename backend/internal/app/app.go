package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/leketech/OpsPilot-AI/backend/internal/agent/orchestrator"
	"github.com/leketech/OpsPilot-AI/backend/internal/config"
)

type Application struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Fiber  *fiber.App
	Agent  *orchestrator.Orchestrator
}

func NewApplication(cfg *config.Config, db *pgxpool.Pool, rdb *redis.Client) *Application {
	fiberApp := fiber.New(fiber.Config{
		AppName: cfg.AppName + " API v0.1.0",
	})
	fiberApp.Use(recover.New())
	fiberApp.Use(cors.New())

	return &Application{
		Config: cfg,
		DB:     db,
		Redis:  rdb,
		Fiber:  fiberApp,
		Agent:  orchestrator.New(cfg, rdb),
	}
}
