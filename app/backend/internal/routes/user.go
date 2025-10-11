package routes

import (
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/handlers"
	"github.com/Quasar777/buildefect/app/backend/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterUserRoutes(app *fiber.App, db *gorm.DB, jwtSecret string) {
	uh := handlers.NewUserHandler(db)
	ah := handlers.NewAuthHandler(db, jwtSecret, 24*time.Hour)

	// auth
	app.Post("/api/auth/register", ah.Register)
	app.Post("/api/auth/login", ah.Login)

	// user managing
	app.Post("/api/users", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRole("observer"), 
		uh.CreateUser,
	)
	app.Get("/api/users", uh.GetUsers)
	app.Get("/api/users/:id", uh.GetUser)
	app.Patch("/api/users/:id", uh.UpdateUser)
	app.Delete("/api/users/:id", uh.DeleteUser)
}