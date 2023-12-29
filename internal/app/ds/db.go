package ds

import (
	"encoding/json"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Region struct {
	ID             uint `gorm:"primaryKey;AUTO_INCREMENT"`
	District       string
	Name           string `gorm:"type:varchar(50);unique;not null"`
	Details        string `gorm:"type:text"`
	Status         string `gorm:"not null"`
	AreaKm         json.Number
	Population     json.Number
	HeadName       string `gorm:"type:varchar(250)"`
	HeadEmail      string `gorm:"type:varchar(50)"`
	HeadPhone      string `gorm:"type:varchar(50)"`
	AverageHeightM json.Number
	ImageName      string
}

type Flight struct {
	ID             uint           `gorm:"primaryKey;AUTO_INCREMENT"`
	ModeratorRefer uuid.UUID      `gorm:"type:uuid"`
	UserRefer      uuid.UUID      `gorm:"type:uuid;not null"`
	Status         string         `gorm:"type:varchar(50)"`
	DateCreated    datatypes.Date `gorm:"not null" swaggertype:"primitive,string"`
	DateProcessed  datatypes.Date `swaggertype:"primitive,string"`
	DateFinished   datatypes.Date `swaggertype:"primitive,string"`
	Moderator      User           `gorm:"foreignKey:ModeratorRefer;references:UUID"`
	User           User           `gorm:"foreignKey:UserRefer;references:UUID;not null"`
	TakeoffDate    datatypes.Date `gorm:"not null" swaggertype:"primitive,string"`
	ArrivalDate    datatypes.Date `gorm:"not null" swaggertype:"primitive,string"`
	AllowedHours   string         `swaggertype:"primitive,string"`
}

type FlightToRegion struct {
	ID          uint   `gorm:"primaryKey;AUTO_INCREMENT"`
	FlightRefer int    `gorm:"not null"`
	RegionRefer int    `gorm:"not null"`
	Flight      Flight `gorm:"foreignKey:FlightRefer"`
	Region      Region `gorm:"foreignKey:RegionRefer"`
}
