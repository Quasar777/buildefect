// DEPRECATED FILE

package database

import (
	"log"
	"os"
	"time"

	"github.com/Quasar777/buildefect/app/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectDbFromDSN(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database\n", err.Error())
		os.Exit(2)
	}

	// получить *sql.DB чтобы настроить пул и ping
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("Failed to get sql.DB from gorm DB: %v", err)
    }

	sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    if err = sqlDB.Ping(); err != nil {
        log.Fatalf("Failed to ping DB: %v", err)
    }

	log.Println("Connected to the db successfully")
	db.Logger = logger.Default.LogMode(logger.Info);
	log.Println("Running migrations")
	// add migrations
	if err := db.AutoMigrate(
		&models.User{}, 
		&models.Building{}, 
		&models.Comment{}, 
		&models.CommentAttachment{},
		&models.Defect{},
		&models.DefectAttachment{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	Database = DbInstance{Db: db}
}