package ds

import "time"

type BookRequestBody struct {
	TakeoffDate string
	ArrivalDate string
	Regions     []string
	Status      string
}

type EditFlightRequestBody struct {
	FlightID    int       `json:"flightID"`
	TakeoffDate time.Time `json:"takeoffDate"`
	ArrivalDate time.Time `json:"arrivalDate"`
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
