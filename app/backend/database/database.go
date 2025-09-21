package database

import (
	"log"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DbInstance struct {
	Db *gorm.DB
}

var Database DbInstance

func ConnectDb() {
	db, err := gorm.Open(sqlite.Open("buildefect.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database\n", err.Error())
		os.Exit(2)
	}

	log.Println("Connected to the db successfully")
	db.Logger = logger.Default.LogMode(logger.Info);
	log.Println("Running migrations")
	// TODO: add migrations

	Database = DbInstance{Db: db}
}