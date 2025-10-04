package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


func RegisterBuildingRoutes(app *fiber.App, db *gorm.DB) {
	h := handlers.NewBuildingHandler(db)

	// TODO: сделать возможность обновления и создания данных только для ролей observer и manager. 
	// Сейчас оставлю так для удобства тестирования

	app.Post("/api/buildings", h.CreateBuilding)
	app.Get("/api/buildings", h.GetBuildings)
	app.Get("/api/buildings/:id", h.GetBuilding)
	app.Patch("/api/buildings/:id", h.UpdateBuilding)
	app.Delete("/api/buildings/:id", h.DeleteBuilding)
}