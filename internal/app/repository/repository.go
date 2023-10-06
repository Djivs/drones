package repository

import (
	"log"
	"strings"

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

	all_regions, all_regions_err := r.GetAllRegions()

	if all_regions_err != nil {
		return nil, all_regions_err
	}

	for i := range all_regions {
		if strings.Contains(strings.ToLower(all_regions[i].Name), strings.ToLower(region_name)) {
			regions = append(regions, all_regions[i])
		}
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

func (r *Repository) FilterActiveRegions(regions []ds.Region) []ds.Region {
	var new_regions = []ds.Region{}

	for i := range regions {
		if regions[i].Status == "Действует" {
			new_regions = append(new_regions, regions[i])
		}
	}

	return new_regions

}

func (r *Repository) LogicalDeleteRegion(region_name string) error {
	return r.db.Model(&ds.Region{}).Where("name = ?", region_name).Update("status", "Недоступен").Error
}

func (r *Repository) ChangeRegionVisibility(region_name string) error {
	region, err := r.GetRegionByName(region_name)

	if err != nil {
		log.Println(err)
		return err
	}

	new_status := ""

	if region.Status == "Действует" {
		new_status = "Недоступен"
	} else {
		new_status = "Действует"
	}

	return r.db.Model(&ds.Region{}).Where("name = ?", region_name).Update("status", new_status).Error
}

func (r *Repository) DeleteRegion(region_name string) error {
	return r.db.Delete(&ds.Region{}, "name = ?", region_name).Error
}

func (r *Repository) CreateRegion(region ds.Region) error {
	return r.db.Create(&region).Error
}

func (r *Repository) CreateUser(user ds.User) error {
	return r.db.Create(&user).Error
}

func (r *Repository) CreateFlight(flight ds.Flight) error {
	return r.db.Create(&flight).Error
}

func (r *Repository) CreateFlightToRegion(flight_to_region ds.FlightToRegion) error {
	return r.db.Create(&flight_to_region).Error
}

func (r *Repository) FindRegion(region ds.Region) (ds.Region, error) {
	var result ds.Region
	err := r.db.Where(&region).First(&result).Error
	if err != nil {
		return ds.Region{}, err
	} else {
		return result, nil
	}
}
