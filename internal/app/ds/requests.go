package ds

type PingRequestBody struct {
	Message string
}

type BookRegionRequestBody struct {
	UserName    string
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
