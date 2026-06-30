package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// Define a clean struct for your JSON response
type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

func main() {
	// 1. Load environment variables from a local .env file (if it exists)
	// We ignore the error because in production (Docker/K8s), env vars are 
	// injected directly by the system and a .env file won't be there.
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env file found, using system environment variables")
	} else {
		log.Info(".env file loaded successfully")
	}

	// 2. Configure Logrus for beautiful, professional logs
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	// 3. App Configuration
	app := fiber.New(fiber.Config{
		AppName: "OpsPilot-AI API v0.1.0",
	})

	// 4. Add Middlewares
	app.Use(recover.New()) // Prevents server crashes on panics
	app.Use(cors.New())    // Allows frontend apps to call this API

	// 5. Define Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(healthResponse{
			Status:    "ok",
			Service:   "OpsPilot AI",
			Version:   "0.1.0",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	// 6. Dynamic Port Configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default fallback
	}

	// 7. Start Server using Logrus instead of standard log
	log.Infof("🚀 starting OpsPilot-AI API on port :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}