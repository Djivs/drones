package app

import (
	"log"
	"net/http"

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

	a.r.PUT("region", a.add_region)

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

func edit_region(c *gin.Context) {

}

func delete_region(c *gin.Context) {

}

func add_to_last_flight(c *gin.Context) {

}

func flights(c *gin.Context) {

}

func flight(c *gin.Context) {

}

func edit_flight(c *gin.Context) {

}

func flight_status_change_creator(c *gin.Context) {

}

func flight_status_change_moderator(c *gin.Context) {

}

func delete_flight(c *gin.Context) {

}

func delete_flight_to_region(c *gin.Context) {

}

func change_flight_to_region_value(c *gin.Context) {

}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
