package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Card struct {
	ID    int
	Title string
	Text  string
	Image string
}

type CardPage struct {
	Title string
	Text  string
	Image string
}

func StartServer() {
	log.Println("Server start up")
	cards := []Card{
		{1, "Беспилотная авиапочта", "", "image/delivery-drone.jpg"},
		{2, "Коммерческая доставка", "", "image/commercial-delivery-drone.jpg"},
		{3, "Дрон-рейсинг", "", "image/drone-race.jpg"},
		{4, "Панорамная съёмка", "", "image/drone-camera.jpg"},
	}
	cardPages := []CardPage{
		{"Беспилотная авиапочта", "Здесь вы сможете создать заявку для беспилотной авиапочты", "../image/delivery-drone.jpg"},
		{"Коммерческая доставка", "Здесь вы сможете создать заявку для коммерческой доставки", "../image/commercial-delivery-drone.jpg"},
		{"Дрон-рейсинг", "Здесь вы сможете создать заявку для дрон-рейсинга", "../image/drone-race.jpg"},
		{"Панорамная съёмка", "Здесь вы сможете создать заявку для панорманой съёмки", "../image/drone-camera.jpg"},
	}

	idToPage := make(map[int]CardPage)
	idToPage[1] = cardPages[0]
	idToPage[2] = cardPages[1]
	idToPage[3] = cardPages[2]
	idToPage[4] = cardPages[3]

	r := gin.Default()

	r.GET("/ping", ping)

	r.LoadHTMLGlob("../../templates/*.html")

	r.Static("/image", "../../resources/image")
	r.Static("/css", "../../templates/css")
	r.Static("/font", "../../resources/font")

	r.GET("/home", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"cards": cards,
		})
	})
	r.GET("/home/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			panic(err)
		}

		cardPage := idToPage[id]

		c.HTML(http.StatusOK, "page.html", gin.H{
			"Title": cardPage.Title,
			"Image": cardPage.Image,
			"Text":  cardPage.Text,
		})
	})
	r.GET("/search", func(c *gin.Context) {
		card_title := c.Query("card_title")

		if card_title == "" {
			c.Redirect(http.StatusFound, "/home")
		}

		foundCards := []Card{}
		for i := range cards {
			if strings.HasPrefix(cards[i].Title, card_title) {
				foundCards = append(foundCards, cards[i])
			}
		}

		c.HTML(http.StatusOK, "search.html", gin.H{
			"cards": foundCards,
		})

	})

	r.Run(":8000")

	log.Println("Server down")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
