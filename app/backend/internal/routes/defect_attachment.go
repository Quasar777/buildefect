package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterDefectAttachmentsRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	h := handlers.NewDefectAttachmentHandler(db)

	app.Post("/api/defects/:id/attachments", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		h.UploadDefectAttachment,
	)

	app.Get("/api/defects/:id/attachments", h.GetDefectAttachments)
	
	app.Get("/api/attachments/:id", h.GetDefectAttachment)

	app.Delete("/api/attachments/:id", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		h.DeleteDefectAttachment,
	)
}