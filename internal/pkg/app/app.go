package app

import (
	"log"
	"net/http"

	"drones/internal/app/dsn"
	"drones/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type PingRequestBody struct {
	Message string
}

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

	a.r.Run(":8000")

	log.Println("Server is down")
}

func regions(c *gin.Context) {

}

func add_region(c *gin.Context) {

}

func get_region(c *gin.Context) {

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
	var requestBody PingRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		// DO SOMETHING WITH THE ERROR
	}

	log.Println(requestBody.Message)

	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
