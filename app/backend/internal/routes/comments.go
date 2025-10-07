package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Обновить комментарии нельзя (Сделано для того, чтобы всегда было видно историю)
// TODO: добавить логгирование, а затем добавить возможность редактирования комментария. 

func RegisterCommentsRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	h := handlers.NewCommentHandler(db)

	app.Post("/api/comments", middleware.JWTMiddleware(jwtSecret), h.CreateComment)
	app.Get("/api/comments", h.GetComments)
	app.Get("/api/comments/:id", h.GetComment)
	app.Delete("/api/comments/:id", middleware.JWTMiddleware(jwtSecret), h.DeleteComment)
}