package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func main() {
	app := fiber.New(fiber.Config{
		AppName: "OpsPilot-AI API",
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(healthResponse{
			Status:    "ok",
			Service:   "opspilot-api",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting OpsPilot-AI API on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
