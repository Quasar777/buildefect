package main

import (
	"github.com/Quasar777/buildefect/app/backend/internal/config"
	"github.com/Quasar777/buildefect/app/backend/internal/database/postgresql"
	"github.com/Quasar777/buildefect/app/backend/internal/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gofiber/swagger"
	_ "github.com/Quasar777/buildefect/app/backend/cmd/api/docs"
)

// @title           buildefect api
// @version         1.0
// @description     this is documentation for buildefect API
// @securityDefinitions.apikey  BearerAuth
// @in                         header
// @name                       Authorization
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

	routes.RegisterUserRoutes(app, pg.GormDB, cfg.JWTSecret)
	routes.RegisterBuildingRoutes(app, pg.GormDB, cfg.JWTSecret)
	routes.RegisterDefectRoutes(app, pg.GormDB, cfg.JWTSecret)
	routes.RegisterCommentsRoutes(app, pg.GormDB, cfg.JWTSecret)
	routes.RegisterDefectAttachmentsRoutes(app, pg.GormDB, cfg.JWTSecret)
	// TODO: create api for comment attachments

	// статическая отдача файлов
	app.Static("/uploads", "internal/uploads")
	
	// swagger
    app.Get("/swagger/*", swagger.HandlerDefault)

    // Запуск приложения
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal().Err(err)
	}
}