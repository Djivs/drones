package app

import (
	"log"
	"net/http"
	"strconv"

	"drones/internal/app/ds"
	"drones/internal/app/dsn"
	"drones/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Application struct {
	repo repository.Repository
	r    *gin.Engine
}

func New() Application {
	app := Application{}

	repo, _ := repository.New(dsn.FromEnv())

	app.repo = *repo

	return app

}

func (a *Application) StartServer() {
	log.Println("Server started")

	a.r = gin.Default()
	a.r.GET("regions", a.get_regions)
	a.r.GET("region/:region", a.get_region)

	a.r.GET("flights", a.get_flights)
	a.r.GET("flight", a.get_flight)

	a.r.PUT("book", a.book_region)

	a.r.PUT("region/add", a.add_region)
	a.r.PUT("region/edit", a.edit_region)
	a.r.PUT("flight/edit", a.edit_flight)
	a.r.PUT("flight/status_change/moderator", a.flight_mod_status_change)
	a.r.PUT("flight/status_change/user", a.flight_user_status_change)

	a.r.PUT("region/delete/:region_name", a.delete_region)
	a.r.PUT("region/delete_restore/:region_name", a.delete_restore_region)
	a.r.PUT("flight/delete/:flight_id", a.delete_flight)

	a.r.DELETE("flight_to_region/delete", a.delete_flight_to_region)

	a.r.Run(":8000")

	log.Println("Server is down")
}

func (a *Application) get_regions(c *gin.Context) {
	var name_pattern = c.Query("name_pattern")
	var district = c.Query("district")
	var status = c.Query("status")

	regions, err := a.repo.GetAllRegions(name_pattern, district, status)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, regions)
}

func (a *Application) add_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil {
		c.String(http.StatusBadRequest, "Can't parse region\n"+err.Error())
		return
	}

	err := a.repo.CreateRegion(region)

	if err != nil {
		c.String(http.StatusNotFound, "Can't create region\n"+err.Error())
		return
	}

	c.String(http.StatusCreated, "Region created successfully")

}

func (a *Application) get_region(c *gin.Context) {
	var region = ds.Region{}
	region.Name = c.Param("region")

	found_region, err := a.repo.FindRegion(region)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, found_region)

}

func (a *Application) edit_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.EditRegion(region)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Region was successfuly edited")

}

func (a *Application) delete_region(c *gin.Context) {
	region_name := c.Param("region_name")

	log.Println(region_name)

	err := a.repo.LogicalDeleteRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region was successfully deleted")
}

func (a *Application) delete_restore_region(c *gin.Context) {
	region_name := c.Param("region_name")

	err := a.repo.DeleteRestoreRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region status was successfully switched")
}

func (a *Application) book_region(c *gin.Context) {
	var request_body ds.BookRegionRequestBody

	if err := c.BindJSON(&request_body); err != nil {
		c.Error(err)
		c.String(http.StatusBadGateway, "Cant' parse json")
		return
	}

	err := a.repo.BookRegion(request_body)

	if err != nil {
		c.Error(err)
		c.String(http.StatusNotFound, "Can't book region")
		return
	}

	c.String(http.StatusCreated, "Region was successfully booked")

}

func (a *Application) get_flights(c *gin.Context) {
	var requestBody ds.GetFlightsRequestBody

	c.BindJSON(&requestBody)

	flights, err := a.repo.GetAllFlights(requestBody)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, flights)
}

func (a *Application) get_flight(c *gin.Context) {
	var flight ds.Flight

	if err := c.BindJSON(&flight); err != nil {
		c.Error(err)
		return
	}

	found_flight, err := a.repo.FindFlight(flight)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, found_flight)
}

func (a *Application) edit_flight(c *gin.Context) {
	var flight ds.Flight

	if err := c.BindJSON(&flight); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.EditFlight(flight)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight was successfuly edited")
}

func (a *Application) flight_mod_status_change(c *gin.Context) {
	var requestBody ds.ChangeFlightStatusRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	user_role, err := a.repo.GetUserRole(requestBody.UserName)

	if err != nil {
		c.Error(err)
		return
	}

	if user_role != "Модератор" {
		c.String(http.StatusBadRequest, "у пользователя должна быть роль модератора")
		return
	}

	err = a.repo.ChangeFlightStatus(requestBody.ID, requestBody.Status)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight status was successfully changed")
}

func (a *Application) flight_user_status_change(c *gin.Context) {
	var requestBody ds.ChangeFlightStatusRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.ChangeFlightStatus(requestBody.ID, requestBody.Status)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight status was successfully changed")
}

func (a *Application) delete_flight(c *gin.Context) {
	flight_id, _ := strconv.Atoi(c.Param("flight_id"))

	err := a.repo.LogicalDeleteFlight(flight_id)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Flight was successfully deleted")
}

func (a *Application) delete_flight_to_region(c *gin.Context) {
	var requestBody ds.DeleteFlightToRegionRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.DeleteFlightToRegion(requestBody.FlightID, requestBody.RegionID)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight-to-region m-m was successfully deleted")
}
