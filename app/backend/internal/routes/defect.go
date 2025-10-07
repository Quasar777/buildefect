package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TODO: сделать ручку для обновления дефекта.

func RegisterDefectRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	h := handlers.NewDefectHandler(db)

	app.Post("/api/defects", middleware.JWTMiddleware(jwtSecret), h.CreateDefect)
	app.Get("/api/defects", h.GetDefects)
	app.Get("/api/defects/:id", h.GetDefect)
	app.Delete("api/defects/:id", h.DeleteDefect)
}