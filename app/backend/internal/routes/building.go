package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


func RegisterBuildingRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	h := handlers.NewBuildingHandler(db)

	app.Post("/api/buildings", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		h.CreateBuilding,
	)

	app.Get("/api/buildings", h.GetBuildings)

	app.Get("/api/buildings/:id", h.GetBuilding)

	app.Patch("/api/buildings/:id", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		h.UpdateBuilding,
	)

	app.Delete("/api/buildings/:id", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		h.DeleteBuilding,
	)
}