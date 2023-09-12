package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Card struct {
	Title   string
	Text    string
	Image   string
	Details string
}

func StartServer() {
	log.Println("Server start up")
	cards := []Card{
		{"Кузьминки", "ЮВАО", "image/kuzminki.jpg", "Заявка для Кузьминок"},
		{"Люберцы", "МО", "image/lyubertsi.jpg", "Заявка для Люберец"},
		{"Савёловский", "САО", "image/savyolovsky.jpg", "Заявка для Савёловского"},
		{"Строгино", "СЗАО", "image/strogino.jpg", "Заявка для Строгино"},
	}

	r := gin.Default()

	r.GET("/ping", ping)

	r.LoadHTMLGlob("../../templates/*.html")

	r.Static("/image", "../../resources/image")
	r.Static("/css", "../../templates/css")
	r.Static("/font", "../../resources/font")

	r.GET("/", func(c *gin.Context) {
		card_title := c.Query("card_title")

		if card_title == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"cards": cards,
			})
			return
		}

		foundCards := []Card{}
		lowerCardTitle := strings.ToLower(card_title)
		for i := range cards {
			if strings.Contains(strings.ToLower(cards[i].Title), lowerCardTitle) {
				foundCards = append(foundCards, cards[i])
			}
		}

		c.HTML(http.StatusOK, "search.html", gin.H{
			"cards":       foundCards,
			"Amount":      len(foundCards),
			"Search_text": card_title,
		})

	})
	r.GET("/:title", func(c *gin.Context) {
		title := c.Param("title")

		for i := range cards {
			if cards[i].Title == title {
				c.HTML(http.StatusOK, "page.html", gin.H{
					"Title":   cards[i].Title,
					"Image":   "../" + cards[i].Image,
					"Text":    cards[i].Text,
					"Details": cards[i].Details,
				})
			}
		}
	})

	r.Run(":8000")

	log.Println("Server down")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
