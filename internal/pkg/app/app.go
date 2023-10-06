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
	a.r.GET("ping", ping)
	a.r.GET("regions", a.get_regions)
	a.r.GET("region", a.get_region)
	a.r.GET("flights", a.get_flights)

	a.r.POST("book", a.book_region)

	a.r.PUT("region/add", a.add_region)
	a.r.PUT("region/edit", a.edit_region)

	a.r.DELETE("region/delete/:region_name", a.delete_region)

	a.r.Run(":8000")

	log.Println("Server is down")
}

func (a *Application) get_regions(c *gin.Context) {
	regions, err := a.repo.GetAllRegions()
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, regions)
}

func (a *Application) add_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.CreateRegion(region)

	if err != nil {
		log.Println(err)
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Region created successfully")

}

func (a *Application) get_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil {
		c.Error(err)
		return
	}

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
	region_name := c.Query("region_name")

	err := a.repo.DeleteRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region was successfully deleted")
}

func (a *Application) book_region(c *gin.Context) {
	var request_body ds.BookRegionRequestBody

	if err := c.BindJSON(&request_body); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.BookRegion(request_body)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Region was successfully booked")

}

func (a *Application) get_flights(c *gin.Context) {
	flights, err := a.repo.GetAllFlights()
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

	c.String(http.StatusCreated, "Region was successfuly edited")
}

func (a *Application) flight_status_change(c *gin.Context) {
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

	err := a.repo.DeleteFlight(flight_id)

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

	c.String(http.StatusCreated, "Flight status was successfully changed")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
