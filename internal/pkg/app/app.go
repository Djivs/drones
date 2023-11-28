package app

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"drones/docs"
	"drones/internal/app/config"
	"drones/internal/app/ds"
	"drones/internal/app/dsn"
	"drones/internal/app/redis"
	"drones/internal/app/repository"
	"drones/internal/app/role"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
)

// @BasePath /

type Application struct {
	repo   *repository.Repository
	r      *gin.Engine
	config *config.Config
	redis  *redis.Client
}

type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func New(ctx context.Context) (*Application, error) {
	cfg, err := config.NewConfig(ctx)
	if err != nil {
		return nil, err
	}

	repo, err := repository.New(dsn.FromEnv())
	if err != nil {
		return nil, err
	}

	redisClient, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, err
	}

	return &Application{
		config: cfg,
		repo:   repo,
		redis:  redisClient,
	}, nil
}

func (a *Application) StartServer() {
	log.Println("Server started")

	a.r = gin.Default()
	a.r.GET("regions", a.get_regions)
	a.r.GET("region/:region", a.get_region)

	a.r.GET("flight", a.get_flight)

	a.r.PUT("book", a.book_region)

	a.r.PUT("region/edit", a.edit_region)
	a.r.PUT("flight/edit", a.edit_flight)
	a.r.PUT("flight/status_change/moderator", a.flight_mod_status_change)
	a.r.PUT("flight/status_change/user", a.flight_user_status_change)

	a.r.PUT("region/delete/:region_name", a.delete_region)
	a.r.PUT("region/delete_restore/:region_name", a.delete_restore_region)
	a.r.PUT("flight/delete/:flight_id", a.delete_flight)
	a.r.PUT("flight_to_region/delete", a.delete_flight_to_region)

	// swagger
	docs.SwaggerInfo.BasePath = "/"
	a.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// registration & etc
	a.r.POST("/login", a.login)
	a.r.POST("/register", a.register)
	a.r.POST("/logout", a.logout)
	a.r.Use(a.WithAuthCheck(role.Moderator, role.Admin, role.User)).GET("flights", a.get_flights)
	a.r.Use(a.WithAuthCheck(role.Moderator, role.Admin)).PUT("region/add", a.add_region)
	a.r.Use(a.WithAuthCheck(role.Admin)).GET("/ping", a.Ping)

	a.r.Run(":8000")

	log.Println("Server is down")
}

// @Summary Get all existing regions
// @Description Returns all existing regions
// @Tags regions
// @Accept json
// @Produce json
// @Success 200 {} json
// @Param name_pattern query string false "Regions name pattern"
// @Param district query string false "Regions district"
// @Param status query string false "Regions status (Действует/Недействителен)"
// @Router /regions [get]
func (a *Application) get_regions(c *gin.Context) {
	var name_pattern = c.Query("name_pattern")
	var district = c.Query("district")
	var status = c.Query("status")

	regions, err := a.repo.GetAllRegions(name_pattern, district, status)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, regions)
}

// @Summary      Adds region to database
// @Description  Creates a new reigon with parameters, specified in json
// @Tags regions
// @Accept json
// @Produce      json
// @Param region body ds.Region true "New region's details"
// @Success      201  {object}  string "Region created successfully"
// @Router       /region/add [put]
func (a *Application) add_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil || region.Name == "" || region.Status == "" {
		c.String(http.StatusBadRequest, "Can't parse region\n"+err.Error())
		return
	}

	err := a.repo.CreateRegion(region)

	if err != nil {
		c.String(http.StatusNotFound, "Can't create region\n"+err.Error())
		return
	}

	c.String(http.StatusCreated, "Region created successfully")

}

// @Summary      Get region
// @Description  Returns region with given name
// @Tags         regions
// @Produce      json
// @Param region path string true "Regions name"
// @Success      200  {object}  string
// @Router       /region/{region} [get]
func (a *Application) get_region(c *gin.Context) {
	var region = ds.Region{}
	region.Name = c.Param("region")

	found_region, err := a.repo.FindRegion(region)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, found_region)

}

// @Summary      Edits region
// @Description  Finds region by name and updates its fields
// @Tags         regions
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param region body ds.Region true "Edited regioons data (must contain regions name or id)"
// @Router       /region/edit [put]
func (a *Application) edit_region(c *gin.Context) {
	var region *ds.Region

	if err := c.BindJSON(&region); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.EditRegion(region)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Region was successfuly edited")

}

// @Summary      Deletes region
// @Description  Finds region by name and changes its status to "Недоступен"
// @Tags         regions
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param region_name path string true "Regions name"
// @Router       /region/delete/{region_name} [put]
func (a *Application) delete_region(c *gin.Context) {
	region_name := c.Param("region_name")

	if region_name == "" {
		c.String(http.StatusBadRequest, "You must specify region name")

		return
	}

	log.Println(region_name)

	err := a.repo.LogicalDeleteRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region was successfully deleted")
}

// @Summary      Deletes or restores region
// @Description  Switches region status from "Действует" to "Недоступен" and back
// @Tags         regions
// @Produce      json
// @Success      200  {object}  string
// @Param region_name path string true "Regions name"
// @Router       /region/delete_restore/{region_name} [get]
func (a *Application) delete_restore_region(c *gin.Context) {
	region_name := c.Param("region_name")

	if region_name == "" {
		c.String(http.StatusBadRequest, "You must specify region name")
	}

	err := a.repo.DeleteRestoreRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Region status was successfully switched")
}

// @Summary      Book region
// @Description  Creates a new flight and adds current region in it
// @Tags general
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param Body body ds.BookRegionRequestBody true "Booking request parameters"
// @Router       /book [put]
func (a *Application) book_region(c *gin.Context) {
	var request_body ds.BookRegionRequestBody

	if err := c.BindJSON(&request_body); err != nil {
		c.Error(err)
		c.String(http.StatusBadGateway, "Cant' parse json")
		return
	}

	err := a.repo.BookRegion(request_body)

	if err != nil {
		c.Error(err)
		c.String(http.StatusNotFound, "Can't book region")
		return
	}

	c.String(http.StatusCreated, "Region was successfully booked")

}

// @Summary      Get flights
// @Description  Returns list of all available flights
// @Tags         flights
// @Produce      json
// @Success      302  {object}  string
// @Param status query string false "Flights status"
// @Router       /flights [get]
func (a *Application) get_flights(c *gin.Context) {
	_roleNumber, _ := c.Get("role")
	_userUUID, _ := c.Get("userUUID")

	roleNumber := _roleNumber.(role.Role)
	userUUID := _userUUID.(uuid.UUID)

	status := c.Query("status")

	flights, err := a.repo.GetAllFlights(status, roleNumber, userUUID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, flights)
}

// @Summary      Get flight
// @Description  Returns flight with given parameters
// @Tags         flights
// @Accept		 json
// @Produce      json
// @Success      302  {object}  string
// @Param status query string false "Flights status"
// @Router       /flight [get]
func (a *Application) get_flight(c *gin.Context) {
	status := c.Query("status")

	flight := &ds.Flight{
		Status: status,
	}

	if err := c.BindJSON(&flight); err != nil {
		c.Error(err)
		return
	}

	found_flight, err := a.repo.FindFlight(flight)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusFound, found_flight)
}

// @Summary      Edits flight
// @Description  Finds flight and updates it fields
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param flight body ds.Flight false "Flight"
// @Router       /flight/edit [put]
func (a *Application) edit_flight(c *gin.Context) {
	var flight *ds.Flight

	if err := c.BindJSON(flight); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.EditFlight(flight)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight was successfuly edited")
}

// @Summary      Changes flight status as moderator
// @Description  Changes flight status to any available status
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param request_body body ds.ChangeFlightStatusRequestBody true "Request body"
// @Router       /flight/status_change/moderator [put]
func (a *Application) flight_mod_status_change(c *gin.Context) {
	var requestBody ds.ChangeFlightStatusRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	user_role, err := a.repo.GetUserRole(requestBody.UserName)

	if err != nil {
		c.Error(err)
		return
	}

	if user_role != role.Moderator {
		c.String(http.StatusBadRequest, "у пользователя должна быть роль модератора")
		return
	}

	err = a.repo.ChangeFlightStatus(requestBody.ID, requestBody.Status)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight status was successfully changed")
}

// @Summary      Changes flights status as user
// @Description  Changes flight status as allowed to user
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param request_body body ds.ChangeFlightStatusRequestBody true "Request body"
// @Router       /flight/status_change/user [put]
func (a *Application) flight_user_status_change(c *gin.Context) {
	var requestBody ds.ChangeFlightStatusRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.ChangeFlightStatus(requestBody.ID, requestBody.Status)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight status was successfully changed")
}

// @Summary      Deletes flight
// @Description  Changes flight status to "Удалён"
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param flight_id path int true "Flight id"
// @Router       /flight/delete/{flight_id} [put]
func (a *Application) delete_flight(c *gin.Context) {
	flight_id, _ := strconv.Atoi(c.Param("flight_id"))

	err := a.repo.LogicalDeleteFlight(flight_id)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Flight was successfully deleted")
}

// @Summary      Deletes flight_to_region connection
// @Description  Deletes region from flight
// @Tags         flights
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param request_body body ds.DeleteFlightToRegionRequestBody true "Request body"
// @Router       /flight_to_region/delete [put]
func (a *Application) delete_flight_to_region(c *gin.Context) {
	var requestBody ds.DeleteFlightToRegionRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	err := a.repo.DeleteFlightToRegion(requestBody.FlightID, requestBody.RegionID)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Flight-to-region m-m was successfully deleted")
}

type pingReq struct{}
type pingResp struct {
	Status string `json:"status"`
}

// @Summary      Show hello text
// @Description  very very friendly response
// @Tags         Tests
// @Produce      json
// @Success      200  {object}  pingResp
// @Router       /ping/{name} [get]
func (a *Application) Ping(gCtx *gin.Context) {
	name := gCtx.Param("name")
	gCtx.String(http.StatusOK, "Hello %s", name)
}

// @Summary Login into system
// @Description Returns your token
// @Tags auth
// @Produce json
// @Accept json
// @Success 200 {object} loginResp
// @Param request_body body loginReq true "Login request body"
// @Router /login [post]
func (a *Application) login(c *gin.Context) {
	req := &loginReq{}

	err := json.NewDecoder(c.Request.Body).Decode(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)

		return
	}

	log.Println(req.Login)

	user, err := a.repo.GetUserByLogin(req.Login)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	log.Println(user)

	if req.Login == user.Name && user.Pass == generateHashString(req.Password) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(3600000000000).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "dj1vs",
			},
			UserUUID: uuid.New(), // test uuid
			Scopes:   []string{}, // test data
			Role:     user.Role,
		})

		if token == nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("token is nil"))

			return
		}

		jwtToken := "test"

		strToken, err := token.SignedString([]byte(jwtToken))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant read str token"))

			return
		}

		c.SetCookie("drones-api-token", "Bearer "+strToken, 3600000000000, "", "", true, true)

		c.JSON(http.StatusOK, loginResp{
			ExpiresIn:   3600000000000,
			AccessToken: strToken,
			TokenType:   "Bearer",
		})
	}

	c.AbortWithStatus(http.StatusForbidden)
}

type registerReq struct {
	Name string `json:"name"` // лучше назвать то же самое что login
	Pass string `json:"pass"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

// @Summary register a new user
// @Description adds a new user to the database
// @Tags auth
// @Produce json
// @Accept json
// @Success 200 {object} registerResp
// @Param request_body body registerReq true "Request body"
// @Router /register [post]
func (a *Application) register(c *gin.Context) {
	req := &registerReq{}
	err := json.NewDecoder(c.Request.Body).Decode(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if req.Pass == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Password should not be empty"))
		return
	}
	if req.Name == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Name should not be empty"))
	}

	err = a.repo.Register(&ds.User{
		UUID: uuid.New(),
		Role: role.User,
		Name: req.Name,
		Pass: generateHashString(req.Pass),
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

// @Summary Logout
// @Details Deactivates user's current token
// @Tags auth
// @Produce json
// @Accept json
// @Success 200
// @Router /logout [post]
func (a *Application) logout(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		c.AbortWithStatus(http.StatusBadRequest)

		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test"), nil
	})
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	err = a.redis.WriteJWTToBlackList(c.Request.Context(), jwtStr, 3600000000000)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)

		return
	}

	c.Status(http.StatusOK)
}

func generateHashString(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func createSignedTokenString() (string, error) {
	privateKey, err := ioutil.ReadFile("demo.rsa")
	if err != nil {
		return "", fmt.Errorf("error reading private key file: %v\n", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return "", fmt.Errorf("error parsing RSA private key: %v\n", err)
	}

	token := jwt.New(jwt.SigningMethodRS256)
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("error signing token: %v\n", err)
	}

	return tokenString, nil
}
