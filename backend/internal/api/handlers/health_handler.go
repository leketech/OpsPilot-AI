package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"github.com/leketech/OpsPilot-AI/backend/internal/app"
)

// Health reports liveness info, including downstream connectivity.
func Health(a *app.Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		redisStatus := "ok"
		if _, err := a.Redis.Ping(context.Background()).Result(); err != nil {
			redisStatus = "disconnected"
		}

		return c.JSON(fiber.Map{
			"status":       "ok",
			"db_connected": a.DB != nil,
			"redis_status": redisStatus,
			"app_name":     a.Config.AppName,
		})
	}
}
