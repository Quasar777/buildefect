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

	app.Get("/api/me", middleware.JWTMiddleware(jwtSecret), func(c *fiber.Ctx) error {
		return uh.GetUserByCtx(c) // implement helper in UserHandler to read c.Locals("user_id")
	})

	// user managing
	app.Post("/api/users", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer"), 
		uh.CreateUser,
	)
	
	app.Get("/api/users", 
		middleware.JWTMiddleware(jwtSecret),
		// middleware.RequireRoles("observer"), 
		uh.GetUsers,
	)

	app.Get("/api/users/:id", 
		middleware.JWTMiddleware(jwtSecret),
		// middleware.RequireRoles("observer"), 
		uh.GetUser,
	)

	app.Patch("/api/users/:id", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer"), 
		uh.UpdateUser,
	)

	app.Delete("/api/users/:id", 
		middleware.JWTMiddleware(jwtSecret),
		middleware.RequireRoles("observer"), 
		uh.DeleteUser,
	)
}