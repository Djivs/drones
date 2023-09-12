package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type Card struct {
	Title      string
	Text       string
	Image      string
	Details    string
	AreaKm     float64
	Population int
}

var cards = []Card{
	{"Кузьминки", "ЮВАО", "image/kuzminki.jpg", "Район с красивым парком и добрыми людьми", 8.15, 140670},
	{"Люблино", "ЮВАО", "image/lyublino.jpg", "Район с прекрасным Люблинским парком и богатым культурным наследием", 17.41, 181447},
	{"Савёловский", "САО", "image/savyolovsky.jpg", "Район, в котором располагается Савёловский вокзал", 2.7, 60070},
	{"Строгино", "СЗАО", "image/strogino.jpg", "Район Москвы с красивыми водоёмами и живописными берегами", 16.84, 161980},
}

func StartServer() {
	log.Println("Server start up")

	r := gin.Default()

	r.LoadHTMLGlob("../../templates/*.html")
	r.Static("/image", "../../resources/image")
	r.Static("/css", "../../templates/css")
	r.Static("/font", "../../resources/font")

	r.GET("/", loadHome)
	r.GET("/:title", loadPage)

	r.Run(":8000")

	log.Println("Server down")
}

func loadHome(c *gin.Context) {
	card_title := c.Query("card_title")

	if card_title == "" {
		c.HTML(http.StatusOK, "regions.html", gin.H{
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

	c.HTML(http.StatusOK, "regions.html", gin.H{
		"cards": foundCards,
	})
}

func loadPage(c *gin.Context) {
	title := c.Param("title")

	for i := range cards {
		if cards[i].Title == title {
			c.HTML(http.StatusOK, "region.html", gin.H{
				"Title":      cards[i].Title,
				"Image":      "../" + cards[i].Image,
				"Text":       cards[i].Text,
				"Details":    cards[i].Details,
				"AreaKm":     cards[i].AreaKm,
				"Population": cards[i].Population,
			})
			return
		}
	}
}
