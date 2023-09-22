package ds

import (
	"gorm.io/datatypes"
)

type District struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type RegionStatus struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type Region struct {
	ID                uint `gorm:"primaryKey"`
	DistrictRefer     int
	RegionStatusRefer int
	Name              string
	Details           string
	Status            RegionStatus `gorm:"foreignKey:RegionStatusRefer"`
	District          District     `gorm:"foreignKey:DistrictRefer"`
	AreaKm            float64
	Population        int
	HeadName          string
	HeadEmail         string
	HeadPhone         string
	AverageHeightM    float64
	Image             string `gorm:"type:bytea"`
}

type Role struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type FlightStatus struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

type User struct {
	ID            uint `gorm:"primaryKey"`
	Name          string
	UserRoleRefer int
	Role          Role `gorm:"foreignKey:UserRoleRefer"`
}

type Flight struct {
	ID                uint `gorm:"primaryKey"`
	UserRefer         int
	FlightStatusRefer int
	Status            FlightStatus `gorm:"foreignKey:FlightStatusRefer"`
	DateCreated       datatypes.Date
	DateProcessed     datatypes.Date
	DateFinished      datatypes.Date
	Moderator         User `gorm:"foreignKey:UserRefer"`
	TakeoffDate       datatypes.Date
	ArrivalDate       datatypes.Date
}

type FlightToRegion struct {
	ID          uint `gorm:"primaryKey"`
	FlightRefer int
	RegionRefer int
	Flight      Flight `gorm:"foreignKey:FlightRefer"`
	Region      Region `gorm:"foreignKey:RegionRefer"`
}
