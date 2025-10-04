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

	// protected example: get current user's profile
	app.Get("/api/me", middleware.JWTMiddleware(jwtSecret), func(c *fiber.Ctx) error {
		return uh.GetUserByCtx(c) // implement helper in UserHandler to read c.Locals("user_id")
	})

	// only user with role "observer" can create user with selected role
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