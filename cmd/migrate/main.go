package main

import (
	"drones/internal/app/ds"
	"drones/internal/app/dsn"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	_ = godotenv.Load()
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		panic("failed to connect database! :(")
	}

	// Migrate the schema

	MigrateSchema(db)
}

func MigrateSchema(db *gorm.DB) {
	err := db.AutoMigrate(&ds.User{})
	err = db.AutoMigrate(&ds.Region{})
	err = db.AutoMigrate(&ds.Flight{})
	err = db.AutoMigrate(&ds.FlightToRegion{})

	if err != nil {
		panic(err)
	}
}
