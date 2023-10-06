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
