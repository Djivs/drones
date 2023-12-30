package ds

import "gorm.io/datatypes"

type BookRequestBody struct {
	TakeoffDate string
	ArrivalDate string
	Regions     []string
	Status      string
}

type EditFlightRequestBody struct {
	FlightID    int            `json:"flightID"`
	TakeoffDate datatypes.Date `json:"takeoffDate"`
	ArrivalDate datatypes.Date `json:"arrivalDate"`
	Status      string         `json:"status"`
}

type SetFlightRegionsRequestBody struct {
	FlightID int
	Regions  []string
}

type ChangeFlightStatusRequestBody struct {
	ID     int
	Status string
}

type DeleteFlightToRegionRequestBody struct {
	FlightID int
	RegionID int
}
