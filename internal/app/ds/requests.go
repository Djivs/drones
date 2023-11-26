package ds

type GetRegionsRequestBody struct {
	District string
	Status   string
}

type BookRegionRequestBody struct {
	UserName    string
	TakeoffDate string
	ArrivalDate string
	RegionName  string
}

type ChangeFlightStatusRequestBody struct {
	ID       int
	Status   string
	UserName string
}

type DeleteFlightToRegionRequestBody struct {
	FlightID int
	RegionID int
}
