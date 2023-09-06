package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Card struct {
	Title       string
	Text        string
	Button_text string
	Image       string
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.LoadHTMLGlob("../../templates/*.html")
	r.Static("/image", "../../resources")
	r.Static("/node_modules", "../../node_modules")

	cards := []Card{
		{"Title 1", "Text 1", "Button_text 1", "image/image.jpg"},
		{"Title 2", "Text 2", "Button_text 2", "image/image.jpg"},
		{"Title 3", "Text 3", "Button_text 3", "image/image.jpg"},
		{"Title 4", "Text 4", "Button_text 4", "image/image.jpg"},
	}

	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main Page",
			"cards": cards,
		})
	})

	r.Run()

	log.Println("Server down")
}
