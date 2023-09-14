package repository

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"drones/internal/app/ds"
)

type Repository struct {
	db *gorm.DB
}

func New(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
}

func (r *Repository) GetRegionByID(id int) (*ds.Region, error) {
	region := &ds.Region{}

	err := r.db.First(region, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (r *Repository) CreateRegion(region ds.Region) error {
	return r.db.Create(region).Error
}
