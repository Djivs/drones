package repository

import (
	"time"

	"gorm.io/datatypes"
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

func (r *Repository) GetRegionByName(name string) (*ds.Region, error) {
	region := &ds.Region{}

	err := r.db.First(region, "name = ?", name).Error
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (r *Repository) GetRegionByID(id int) (*ds.Region, error) {
	region := &ds.Region{}

	err := r.db.First(region, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return region, nil
}

func (r *Repository) GetUserByID(id int) (*ds.User, error) {
	user := &ds.User{}

	err := r.db.First(user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserID(name string) (int, error) {
	user := &ds.User{}

	err := r.db.First(user, "name = ?", name).Error
	if err != nil {
		return -1, err
	}

	return int(user.ID), nil
}

func (r *Repository) GetRegionID(name string) (int, error) {
	region := &ds.Region{}

	err := r.db.First(region, "name = ?", name).Error
	if err != nil {
		return -1, err
	}

	return int(region.ID), nil
}

func (r *Repository) GetUserRole(name string) (string, error) {
	user := &ds.User{}

	err := r.db.First(user, "name = ?", name).Error
	if err != nil {
		return "", err
	}

	return user.Role, nil
}

func (r *Repository) GetAllRegions(name_pattern string, district string, status string) ([]ds.Region, error) {
	regions := []ds.Region{}

	var tx *gorm.DB = r.db

	if name_pattern != "" {
		tx = tx.Where("name like ?", "%"+name_pattern+"%")

	}

	if district != "" {
		tx = tx.Where("district = ?", district)
	}
	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	err := tx.Find(&regions).Error

	if err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *Repository) GetAllFlights(requestBody ds.GetFlightsRequestBody) ([]ds.Flight, error) {
	flights := []ds.Flight{}

	var tx *gorm.DB = r.db
	if requestBody.Status != "" {
		tx = tx.Where("status = ?", requestBody.Status)
	}

	err := tx.Find(&flights).Error

	if err != nil {
		return nil, err
	}

	for i := range flights {
		if flights[i].ModeratorRefer != 0 {
			moderator, _ := r.GetUserByID(flights[i].ModeratorRefer)
			flights[i].Moderator = *moderator
		}
		user, _ := r.GetUserByID(flights[i].UserRefer)
		flights[i].User = *user
	}

	return flights, nil
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

func (r *Repository) DeleteRegion(region_name string) error {
	return r.db.Delete(&ds.Region{}, "name = ?", region_name).Error
}

func (r *Repository) DeleteFlight(id int) error {
	return r.db.Delete(&ds.Flight{}, "id = ?", id).Error
}

func (r *Repository) DeleteFlightToRegion(flight_id int, region_id int) error {
	return r.db.Where("flight_refer = ?", flight_id).Where("region_refer = ?", region_id).Delete(&ds.FlightToRegion{}).Error
}

func (r *Repository) LogicalDeleteRegion(region_name string) error {
	return r.db.Model(&ds.Region{}).Where("name = ?", region_name).Update("status", "Недоступен").Error
}

func (r *Repository) LogicalDeleteFlight(flight_id int) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", flight_id).Update("status", "Удалён").Error
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

func (r *Repository) FindFlight(flight ds.Flight) (ds.Flight, error) {
	var result ds.Flight
	err := r.db.Where(&flight).First(&result).Error
	if err != nil {
		return ds.Flight{}, err
	}

	var user ds.User
	r.db.Where("id = ?", result.UserRefer).First(&user)

	result.User = user

	var moderator ds.User
	r.db.Where("id = ?", result.ModeratorRefer).First(&user)

	result.Moderator = moderator

	return result, nil
}

func (r *Repository) EditRegion(region ds.Region) error {
	return r.db.Model(&ds.Region{}).Where("name = ?", region.Name).Updates(region).Error
}

func (r *Repository) EditFlight(flight ds.Flight) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", flight.ID).Updates(flight).Error
}

func (r *Repository) BookRegion(requestBody ds.BookRegionRequestBody) error {
	user_id, err := r.GetUserID(requestBody.UserName)

	if err != nil {
		return err
	}

	var region_id int
	region_id, err = r.GetRegionID(requestBody.RegionName)
	if err != nil {
		return err
	}

	current_date := datatypes.Date(time.Now())
	takeoff_date, err := time.Parse(time.RFC3339, requestBody.TakeoffDate+"T00:00:00Z")
	if err != nil {
		return err
	}
	arrival_date, err := time.Parse(time.RFC3339, requestBody.ArrivalDate+"T00:00:00Z")
	if err != nil {
		return err
	}

	flight := ds.Flight{}
	flight.TakeoffDate = datatypes.Date(takeoff_date)
	flight.ArrivalDate = datatypes.Date(arrival_date)
	flight.UserRefer = user_id
	flight.DateCreated = current_date

	err = r.db.Omit("moderator_refer", "date_processed", "date_finished").Create(&flight).Error

	if err != nil {
		return err
	}

	flight_to_region := ds.FlightToRegion{}
	flight_to_region.FlightRefer = int(flight.ID)
	flight_to_region.RegionRefer = int(region_id)
	err = r.CreateFlightToRegion(flight_to_region)

	return err

}

func (r *Repository) ChangeFlightStatus(id int, status string) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", id).Update("status", status).Error
}
