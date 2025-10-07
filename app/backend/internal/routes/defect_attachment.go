package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterDefectAttachmentsRoutes(app *fiber.App, db *gorm.DB) {
	h := handlers.NewDefectAttachmentHandler(db)

	app.Post("/api/defects/:id/attachments", h.UploadDefectAttachment)
	app.Get("/api/defects/:id/attachments", h.GetDefectAttachments)
	app.Get("/api/attachments/:id", h.GetDefectAttachment)
	app.Delete("/api/attachments/:id", h.DeleteDefectAttachment)
}