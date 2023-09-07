package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Card struct {
	ID          int
	Title       string
	Text        string
	Button_text string
	Image       string
}

type CardPage struct {
	Title string
	Text  string
	Image string
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
		{1, "Title 1", "Text 1", "Button_text 1", "image/image.jpg"},
		{2, "Title 2", "Text 2", "Button_text 2", "image/image.jpg"},
		{3, "Title 3", "Text 3", "Button_text 3", "image/image.jpg"},
		{4, "Title 4", "Text 4", "Button_text 4", "image/image.jpg"},
	}

	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main Page",
			"cards": cards,
		})
	})

	cardPages := []CardPage{
		{"Page title 1", "Page text 1", "../image/image.jpg"},
		{"Page title 2", "Page text 2", "../image/image.jpg"},
		{"Page title 3", "Page text 3", "../image/image.jpg"},
		{"Page title 4", "Page text 4", "../image/image.jpg"},
	}

	r.GET("/home/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}

		cardPage := cardPages[id-1]

		c.HTML(http.StatusOK, "page.html", gin.H{
			"Title": cardPage.Title,
			"Image": cardPage.Image,
			"Text":  cardPage.Text,
		})
	})

	r.Run()

	log.Println("Server down")
}
