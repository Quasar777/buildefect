package postgresql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/config"
	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Database struct {
	GormDB *gorm.DB
	DB *sql.DB
}

func Connect(cfg *config.Config, l zerolog.Logger) (*Database, error) {
	dsn := cfg.DBConnString()

	// GORM logger — используем стандартный, можно тонко настраивать
	gormLog := gormlogger.Default.LogMode(gormlogger.Info)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		l.Error().Err(err).Msg("failed to open gorm DB")
		return nil, fmt.Errorf("failed to open gorm DB: %w", err)
	}

	// Получаем sql.DB для управления пулом и ping
	sqlDB, err := gormDB.DB()
	if err != nil {
		l.Error().Err(err).Msg("failed to get sql.DB from gorm DB")
		return nil, fmt.Errorf("failed to get sql.DB from gorm DB: %w", err)
	}

	// Настройка пула соединений — значения можно вынести в конфиг
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Ping DB, чтобы убедиться, что соединение действительно доступно
	if err := sqlDB.Ping(); err != nil {
		l.Error().Err(err).Msg("failed to ping database")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// запускаем миграции
	if err := gormDB.AutoMigrate(
		&models.User{},
		&models.Building{},
		&models.Comment{},
		&models.CommentAttachment{},
		&models.Defect{},
		&models.DefectAttachment{},
	); err != nil {
		l.Error().Err(err).Msg("auto-migrate failed")
		return nil, fmt.Errorf("auto-migrate failed: %w", err)
	}
	l.Info().Msg("auto-migrate completed")


	l.Info().Msg("connected to Postgres successfully")

	return &Database{
		GormDB: gormDB,
		DB:  sqlDB,
	}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}