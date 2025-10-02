package main

import (
	"github.com/Quasar777/buildefect/app/backend/internal/config"
	"github.com/Quasar777/buildefect/app/backend/internal/database/postgresql"
	"github.com/Quasar777/buildefect/app/backend/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)


func main() {
	// Инициализация логгера
	zerolog.TimeFieldFormat = "02.01.2006 15:04:05.000"
	logger := log.With().Logger()

	// Инициализация конфига
	cfg := config.LoadConfig(logger)

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
	
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		logger.Info().Msg("Request get")
		return c.SendString("Hello, World!")
	})

	app.Post("/", func(c *fiber.Ctx) error {
		logger.Info().Msg("Request post")
		if err := pg.DB.Ping(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusOK).SendString("Pong")
	})

	routes.RegisterUserRoutes(app, pg.GormDB)
	
    // Запуск приложения
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal().Err(err)
	}

}