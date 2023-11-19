package app

import (
	"log"
	"net/http"
	"strconv"

	"drones/internal/app/ds"
	"drones/internal/app/dsn"
	"drones/internal/app/repository"

	docs "drones/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @BasePath /

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
	a.r.PUT("flight_to_region/delete", a.delete_flight_to_region)

	docs.SwaggerInfo.BasePath = "/"
	a.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	a.r.Run(":8000")

	log.Println("Server is down")
}

// @Summary Get all existing regions
// @Schemes
// @Description Returns all existing regions
// @Tags regions
// @Accept json
// @Produce json
// @Success 200 {} string
// @Router /regions [get]
func (a *Application) get_regions(c *gin.Context) {
	var name_pattern = c.Query("name_pattern")
	var district = c.Query("district")
	var status = c.Query("status")

	regions, err := a.repo.GetAllRegions(name_pattern, district, status)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, regions)
}

// @Summary      Adds region to database
// @Description  Creates a new reigon with parameters, specified in json
// @Tags regions
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Router       /region/add [put]
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

// @Summary      Get region
// @Description  Returns region with given name
// @Tags         regions
// @Produce      json
// @Success      200  {object}  string
// @Router       /region/:region [get]
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

// @Summary      Edits region
// @Description  Finds region by name and updates its fields
// @Tags         regions
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Router       /region/edit [put]
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

// @Summary      Deletes region
// @Description  Finds region by name and changes its status to "Недоступен"
// @Tags         regions
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Router       /region/delete/:region_name [put]
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

// @Summary      Deletes or restores region
// @Description  Switches region status from "Действует" to "Недоступен" and back
// @Tags         regions
// @Produce      json
// @Success      200  {object}  string
// @Router       /region/delete_restore/:region_name [get]
func (a *Application) delete_restore_region(c *gin.Context) {
	region_name := c.Param("region_name")

	err := a.repo.DeleteRestoreRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region status was successfully switched")
}

// @Summary      Book region
// @Description  Creates a new flight and adds current region in it
// @Tags general
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Router       /book [put]
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

// @Summary      Get flights
// @Description  Returns list of all available flights
// @Tags         flights
// @Param status string
// @Produce      json
// @Success      302  {object}  string
// @Router       /flights [get]
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

// a.r.GET("flight", a.get_flight)
// @Summary      Get flight
// @Description  Returns flight with given parameters
// @Tags         flights
// @Accept		 json
// @Produce      json
// @Success      302  {object}  string
// @Router       /flight [get]
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

// @Summary      Edits flight
// @Description  Finds flight and updates it fields
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Router       /flight/edit [put]
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

// @Summary      Changes flight status as moderator
// @Description  Changes flight status to any available status
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Router       /flight/status_change/moderator [put]
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

// Ping godoc
// @Summary      Changes flights status as user
// @Description  Changes flight status as allowed to user
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Router       /flight/status_change/user [put]
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

// @Summary      Deletes flight
// @Description  Changes flight status to "Удалён"
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Router       /flight/delete/:flight_id [put]
func (a *Application) delete_flight(c *gin.Context) {
	flight_id, _ := strconv.Atoi(c.Param("flight_id"))

	err := a.repo.LogicalDeleteFlight(flight_id)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Flight was successfully deleted")
}

// @Summary      Deletes flight_to_region connection
// @Description  Deletes region from flight
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Router       /flight_to_region/delete [put]
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
