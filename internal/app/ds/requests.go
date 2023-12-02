package ds

type BookRegionRequestBody struct {
	TakeoffDate string
	ArrivalDate string
	RegionName  string
}

type ChangeFlightStatusRequestBody struct {
	ID     int
	Status string
}

type DeleteFlightToRegionRequestBody struct {
	FlightID int
	RegionID int
}

