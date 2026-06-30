package main

import (
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"

	"github.com/leketech/OpsPilot-AI/backend/internal/api/routes"
	"github.com/leketech/OpsPilot-AI/backend/internal/app"
	"github.com/leketech/OpsPilot-AI/backend/internal/config"
	"github.com/leketech/OpsPilot-AI/backend/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Warn("No .env file found, using system environment variables")
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	cfg := config.Load()

	log.Info("Connecting to PostgreSQL...")
	dbPool, err := database.ConnectPostgres(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Postgres: %v", err)
	}
	defer dbPool.Close() 

	log.Info("Connecting to Redis...")
	redisClient, err := database.ConnectRedis(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// 1. Create the Application Struct
	application := app.NewApplication(cfg, dbPool, redisClient)

	// 2. Register the Routes
	routes.Register(application)

	// 3. Start the Server
	log.Infof("🚀 starting %s on port :%s", cfg.AppName, cfg.Port)
	if err := application.Fiber.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}