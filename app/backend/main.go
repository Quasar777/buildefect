package main

import (
	"github.com/Quasar777/buildefect/app/backend/database"
	"github.com/Quasar777/buildefect/app/backend/database/postgresql"
	"github.com/Quasar777/buildefect/app/backend/internal/config"
	"github.com/Quasar777/buildefect/app/backend/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func setupRoutes(app *fiber.App) {
	// User endpoints
	app.Post("/api/users", routes.CreateUser)
	
	app.Get("/api/users", routes.GetUsers)

	app.Get("/api/users/:id", routes.GetUser)	
	
	app.Put("/api/users/:id", routes.UpdateUser)

	app.Delete("/api/users/:id", routes.DeleteUser)
}

func main() {
	// Инициализация логгера
	zerolog.TimeFieldFormat = "02.01.2006 15:04:05.000"
	logger := log.With().Logger()

	// Инициализация конфига
	cfg := config.LoadConfig(logger)

	// database.ConnectDbFromDSN(cfg.DBConnString())

	// Подключаемся к Postgres
	pg, err := postgresql.Connect(cfg, logger)
	if err != nil {
        logger.Fatal().Err(err).Msg("unable to connect to postgres")
    }
    defer func() {
        if err := pg.Close(); err != nil {
            logger.Error().Err(err).Msg("failed to close db")
        }
    }()

	// Присваиваем GORM DB в глобальную обертку, которую ты используешь в проекте
    database.Database = database.DbInstance{Db: pg.GormDB}
	
	app := fiber.New()

	setupRoutes(app)
	
    // Запуск приложения
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal().Err(err)
	}
	
}