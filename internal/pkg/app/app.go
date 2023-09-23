package app

import (
	"log"
	"net/http"

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

	a.r.LoadHTMLGlob("../../templates/*.html")
	a.r.Static("/css", "../../templates/css")
	a.r.Static("/js", "../../templates/js")

	a.r.GET("/", a.loadRegions)
	a.r.GET("/:region_name", a.loadRegion)
	a.r.POST("/delete_region/:region_name", a.loadRegionChangeVisibility)

	a.r.Run(":8000")

	log.Println("Server is down")
}

func (a *Application) loadRegions(c *gin.Context) {
	region_name := c.Query("region_name")

	if region_name == "" {
		all_regions, err := a.repo.GetAllRegions()

		if err != nil {
			log.Println(err)
			c.Error(err)
		}

		c.HTML(http.StatusOK, "regions.html", gin.H{
			"regions": all_regions,
		})
	} else {
		found_regions, err := a.repo.SearchRegions(region_name)

		if err != nil {
			c.Error(err)
			return
		}

		c.HTML(http.StatusOK, "regions.html", gin.H{
			"regions":     found_regions,
			"Search_text": region_name,
		})
	}
}

func (a *Application) loadRegion(c *gin.Context) {
	region_name := c.Param("region_name")

	if region_name == "favicon.ico" {
		return
	}

	region, err := a.repo.GetRegionByName(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.HTML(http.StatusOK, "region.html", gin.H{
		"Name":           region.Name,
		"Image":          region.Image,
		"AreaKm":         region.AreaKm,
		"Population":     region.Population,
		"Details":        region.Details,
		"HeadName":       region.HeadName,
		"HeadEmail":      region.HeadEmail,
		"HeadPhone":      region.HeadPhone,
		"AverageHeightM": region.AverageHeightM,
		"Region_status":  region.Status,
	})
}

func (a *Application) loadRegionChangeVisibility(c *gin.Context) {
	region_name := c.Param("region_name")
	err := a.repo.ChangeRegionVisibility(region_name)

	if err != nil {
		c.Error(err)
	}

	c.Redirect(http.StatusFound, "/"+region_name)
}
