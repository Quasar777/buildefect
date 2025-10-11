package routes

import (
	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TODO: create endpoint for updating

func RegisterDefectRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	dh := handlers.NewDefectHandler(db)			

	app.Post("/api/defects", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		dh.CreateDefect,
	)
	app.Get("/api/defects", dh.GetDefects)
	app.Get("/api/defects/:id", dh.GetDefect)
	
	app.Delete("api/defects/:id",
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer", "manager"),
		dh.DeleteDefect,
	)
}