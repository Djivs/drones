package repository

import (
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"drones/internal/app/ds"
	"drones/internal/app/role"
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

func (r *Repository) GetUserByID(id uuid.UUID) (*ds.User, error) {
	user := &ds.User{}

	err := r.db.First(user, "UUID = ?", id).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	user := &ds.User{}

	err := r.db.First(user, "name = ?", login).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) GetUserID(name string) (uuid.UUID, error) {
	user := &ds.User{}

	err := r.db.First(user, "name = ?", name).Error
	if err != nil {
		return uuid.Nil, err
	}

	return user.UUID, nil
}

func (r *Repository) GetRegionID(name string) (int, error) {
	region := &ds.Region{}

	err := r.db.First(region, "name = ?", name).Error
	if err != nil {
		return -1, err
	}

	return int(region.ID), nil
}

func (r *Repository) GetRegionStatus(name string) (string, error) {
	region := &ds.Region{}

	err := r.db.First(region, "name = ?", name).Error
	if err != nil {
		return "", err
	}

	return region.Status, nil
}

func (r *Repository) GetUserRole(name string) (role.Role, error) {
	user := &ds.User{}

	err := r.db.First(user, "name = ?", name).Error
	if err != nil {
		return role.Undefined, err
	}

	return user.Role, nil
}

func (r *Repository) GetRegions(name_pattern string, district string, status string) ([]ds.Region, error) {
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

func (r *Repository) GetFlights(status string, startDate string, endDate string, roleNumber role.Role, userUUID uuid.UUID) ([]ds.Flight, error) {
	flights := []ds.Flight{}

	var tx *gorm.DB = r.db
	if status != "" {
		tx = tx.Where("status = ?", status)
	}

	if startDate != "" {
		tx = tx.Where("date_created >= ?", startDate)
	}

	if endDate != "" {
		tx = tx.Where("date_created <= ?", endDate)
	}

	if roleNumber == role.User {
		tx = tx.Where("user_refer = ?", userUUID)
	}

	err := tx.Find(&flights).Error

	if err != nil {
		return nil, err
	}

	log.Println(flights)

	for i := range flights {
		if flights[i].ModeratorRefer != nil {
			moderator, _ := r.GetUserByID(*flights[i].ModeratorRefer)
			flights[i].Moderator = *moderator
		}
		user, _ := r.GetUserByID(*flights[i].UserRefer)
		flights[i].User = *user
	}

	return flights, nil
}

func (r *Repository) GetDraftFlight(user uuid.UUID) (ds.Flight, error) {
	flight := ds.Flight{}

	err := r.db.Where("user_refer = ?", user).Where("status = ?", "Черновик").Find(&flight).Error

	return flight, err
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

func (r *Repository) LogicalDeleteRegion(region_name string) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec(`UPDATE public.regions SET status = ? WHERE name = ?`, "Недоступен", region_name).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) LogicalDeleteFlight(flight_id int) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec(`UPDATE public.flights SET status = ? WHERE id = ?`, "Удалён", flight_id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) ModConfirmFlight(uuid uuid.UUID, flight_id int, confirm bool) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	new_status := "Отклонён"
	if confirm {
		new_status = "Завершён"
	}

	if err := tx.Exec(`UPDATE public.flights SET status = ?, moderator_refer = ?, WHERE id = ?`, new_status, uuid, flight_id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *Repository) UserConfirmFlight(uuid uuid.UUID, flight_id int) error {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Exec(`UPDATE public.flights SET status = ?, user_refer = ?, WHERE id = ?`, "Сформирован", uuid, flight_id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
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

func (r *Repository) FindFlight(flight *ds.Flight) (ds.Flight, error) {
	var result ds.Flight
	err := r.db.Where(&flight).First(&result).Error
	if err != nil {
		return ds.Flight{}, err
	}

	var user ds.User
	r.db.Where("uuid = ?", result.UserRefer).First(&user)

	result.User = user

	var moderator ds.User
	r.db.Where("uuid = ?", result.ModeratorRefer).First(&user)

	result.Moderator = moderator

	return result, nil
}

func (r *Repository) EditRegion(region *ds.Region) error {
	return r.db.Model(&ds.Region{}).Where("name = ?", region.Name).Updates(region).Error
}

func (r *Repository) EditFlight(flight *ds.Flight, moderatorUUID uuid.UUID) error {
	flight.DateProcessed = time.Now()
	flight.ModeratorRefer = &moderatorUUID
	return r.db.Model(&ds.Flight{}).Where("id = ?", flight.ID).Updates(flight).Error
}

func (r *Repository) SetRegionImage(id int, image string) error {
	return r.db.Model(&ds.Region{}).Where("id = ?", id).Update("image_name", image).Error
}

func (r *Repository) Book(requestBody ds.BookRequestBody, userUUID uuid.UUID) error {
	var region_ids []int
	for _, regionName := range requestBody.Regions {
		region_id, err := r.GetRegionID(regionName)
		if err != nil {
			return err
		}
		region_ids = append(region_ids, region_id)
	}

	takeoff_date, err := time.Parse(time.RFC3339, requestBody.TakeoffDate)
	if err != nil {
		return err
	}
	arrival_date, err := time.Parse(time.RFC3339, requestBody.ArrivalDate)
	if err != nil {
		return err
	}

	flight := ds.Flight{}
	flight.TakeoffDate = takeoff_date
	flight.ArrivalDate = arrival_date
	flight.UserRefer = &userUUID
	flight.DateCreated = time.Now()
	flight.Status = requestBody.Status

	err = r.db.Omit("moderator_refer", "date_processed", "date_finished").Create(&flight).Error
	if err != nil {
		return err
	}

	for _, region_id := range region_ids {
		flight_to_region := ds.FlightToRegion{}
		flight_to_region.FlightRefer = int(flight.ID)
		flight_to_region.RegionRefer = int(region_id)
		err = r.CreateFlightToRegion(flight_to_region)

		if err != nil {
			return err
		}
	}

	return nil

}

func (r *Repository) GetFlightStatus(id int) (string, error) {
	var result ds.Flight
	err := r.db.Where("id = ?", id).First(&result).Error
	if err != nil {
		return "", err
	}

	return result.Status, nil
}

func (r *Repository) GetFlightRegions(id int) ([]ds.Region, error) {
	flight_to_regions := []ds.FlightToRegion{}

	err := r.db.Model(&ds.FlightToRegion{}).Where("flight_refer = ?", id).Find(&flight_to_regions).Error
	if err != nil {
		return []ds.Region{}, err
	}

	var regions []ds.Region
	for _, flight_to_region := range flight_to_regions {
		region, err := r.GetRegionByID(flight_to_region.RegionRefer)
		if err != nil {
			return []ds.Region{}, err
		}
		for _, ele := range regions {
			if ele == *region {
				continue
			}
		}
		regions = append(regions, *region)
	}

	return regions, nil

}

func (r *Repository) SetFlightRegions(flightID int, regions []string) error {
	var region_ids []int
	for _, region := range regions {
		region_id, err := r.GetRegionID(region)
		if err != nil {
			return err
		}

		for _, ele := range region_ids {
			if ele == region_id {
				continue
			}
		}

		region_ids = append(region_ids, region_id)
	}

	var existing_links []ds.FlightToRegion
	err := r.db.Model(&ds.FlightToRegion{}).Where("flight_refer = ?", flightID).Find(&existing_links).Error
	if err != nil {
		return err
	}

	for _, link := range existing_links {
		regionFound := false
		regionIndex := -1
		for index, ele := range region_ids {
			if ele == link.RegionRefer {
				regionFound = true
				regionIndex = index
				break
			}
		}

		if regionFound {
			region_ids = append(region_ids[:regionIndex], region_ids[regionIndex+1:]...)
		} else {
			err := r.db.Model(&ds.FlightToRegion{}).Delete(&link).Error
			if err != nil {
				return err
			}
		}
	}

	for _, region_id := range region_ids {
		newLink := ds.FlightToRegion{
			FlightRefer: flightID,
			RegionRefer: region_id,
		}

		err := r.db.Model(&ds.FlightToRegion{}).Create(&newLink).Error
		if err != nil {
			return nil
		}
	}

	return nil
}

func (r *Repository) SetFlightModerator(flightID int, moderatorUUID uuid.UUID) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", flightID).Update("moderator_refer", moderatorUUID).Error
}

func (r *Repository) ChangeFlightStatusUser(id int, status string, userUUID uuid.UUID) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", id).Where("user_refer = ?", userUUID).Update("status", status).Error
}

func (r *Repository) ChangeFlightStatus(id int, status string) error {
	return r.db.Model(&ds.Flight{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) DeleteFlightToRegion(flight_id int, region_id int) error {
	return r.db.Where("flight_refer = ?", flight_id).Where("region_refer = ?", region_id).Delete(&ds.FlightToRegion{}).Error
}

func (r *Repository) Register(user *ds.User) error {
	if user.UUID == uuid.Nil {
		user.UUID = uuid.New()
	}

	return r.db.Create(user).Error
}
