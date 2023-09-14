package app

import (
	"log"
	"net/http"
	"strconv"

	"drones/internal/app/repository"

	"github.com/gin-gonic/gin"
)

type Application struct {
	repo repository.Repository
}

func New() Application {
	return Application{}
}

func (a *Application) StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		id := c.Query("id") // получаем из запроса query string

		if id != "" {
			log.Printf("id recived %s\n", id)
			intID, err := strconv.Atoi(id) // пытаемся привести это к чиселке
			if err != nil {                // если не получилось
				log.Printf("cant convert id %v", err)
				c.Error(err)
				return
			}
			log.Printf("id translated!")

			region, err := a.repo.GetRegionByID(intID)
			if err != nil { // если не получилось
				log.Printf("cant get region by id %v", err)
				c.Error(err)
				return
			}

			log.Printf("got region by id")

			c.JSON(http.StatusOK, gin.H{
				"region_population": region.Population,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(":8000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")

	log.Println("Server down")
}
