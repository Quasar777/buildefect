package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterUserRoutes(app *fiber.App, db *gorm.DB) {
	h := handlers.NewUserHandler(db)

	app.Post("/api/users", h.CreateUser)
	app.Get("/api/users", h.GetUsers)
	app.Get("/api/users/:id", h.GetUser)
	app.Put("/api/users/:id", h.UpdateUser)
	app.Delete("/api/users/:id", h.DeleteUser)
}