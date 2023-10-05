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
