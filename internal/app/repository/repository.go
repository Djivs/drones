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

func (r *Repository) GetRegionByName(name string) (*ds.Region, error) {
	region := &ds.Region{}

	err := r.db.First(region, "name = ?", name).Error
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (r *Repository) SearchRegions(region_name string) ([]ds.Region, error) {
	regions := []ds.Region{}

	err := r.db.Raw("select * from search_region(?)", region_name).Scan(&regions).Error
	if err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *Repository) GetAllRegions() ([]ds.Region, error) {
	regions := []ds.Region{}

	err := r.db.Find(&regions).Error

	if err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *Repository) DeleteRegion(region_name string) error {
	return r.db.Delete(&ds.Region{}, "name = ?", region_name).Error
}

func (r *Repository) CreateRegion(region ds.Region) error {
	return r.db.Create(region).Error
}

func (r *Repository) CreateDistrict(district ds.District) error {
	return r.db.Create(district).Error
}

func (r *Repository) CreateRole(role ds.Role) error {
	return r.db.Create(role).Error
}

func (r *Repository) CreateFlightStatus(flight_status ds.FlightStatus) error {
	return r.db.Create(flight_status).Error
}

func (r *Repository) CreateUser(user ds.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) CreateFlight(flight ds.Flight) error {
	return r.db.Create(flight).Error
}

func (r *Repository) CreateFlightToRegion(flight_to_region ds.FlightToRegion) error {
	return r.db.Create(flight_to_region).Error
}
