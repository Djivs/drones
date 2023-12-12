package ds

type BookRegionRequestBody struct {
	TakeoffDate string
	ArrivalDate string
	RegionName  string
}

type BookRequestBody struct {
	TakeoffDate string
	ArrivalDate string
	Regions     []string
}

type ChangeFlightStatusRequestBody struct {
	ID     int
	Status string
}

type DeleteFlightToRegionRequestBody struct {
	FlightID int
	RegionID int
}
