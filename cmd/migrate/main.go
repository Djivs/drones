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
	MigrateRole(db)
	MigrateFlightStatus(db)
	MigrateUser(db)
	MigrateRegion(db)
	MigrateFlight(db)
	MigrateFlightToRegion(db)
}

func MigrateRegion(db *gorm.DB) {
	err := db.AutoMigrate(&ds.Region{})
	if err != nil {
		panic("cant migrate Region to db")
	}
}

func MigrateRole(db *gorm.DB) {
	err := db.AutoMigrate(&ds.Role{})
	if err != nil {
		panic("cant migrate UserRole to db")
	}
}

func MigrateFlightStatus(db *gorm.DB) {
	err := db.AutoMigrate(&ds.FlightStatus{})
	if err != nil {
		panic("cant migrate FlightStatus to db")
	}
}

func MigrateUser(db *gorm.DB) {
	err := db.AutoMigrate(&ds.User{})
	if err != nil {
		panic("cant migrate User to db")
	}
}

func MigrateFlight(db *gorm.DB) {
	err := db.AutoMigrate(&ds.Flight{})
	if err != nil {
		panic("cant migrate Flight to db")
	}
}

func MigrateFlightToRegion(db *gorm.DB) {
	err := db.AutoMigrate(&ds.FlightToRegion{})
	if err != nil {
		panic("cant migrate FlightToRegion db")
	}
}
