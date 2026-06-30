package routes

import (
	"github.com/leketech/OpsPilot-AI/backend/internal/api/handlers"
	"github.com/leketech/OpsPilot-AI/backend/internal/app"
)

// Register attaches all API routes to the Fiber app.
func Register(a *app.Application) {
	a.Fiber.Get("/health", handlers.Health(a))
}
