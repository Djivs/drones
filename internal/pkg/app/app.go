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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/datatypes"

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
	Login       string `json:"login"`
	Role        int    `json:"role"`
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

	// swagger
	docs.SwaggerInfo.BasePath = "/"
	a.r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	a.r.GET("regions", a.get_regions)
	a.r.GET("region/:region", a.get_region)

	// registration & etc
	a.r.POST("/login", a.login)
	a.r.POST("/register", a.register)
	a.r.POST("/logout", a.logout)

	a.r.Use(a.WithAuthCheck(role.Moderator, role.Admin, role.User)).GET("flight", a.get_flight)
	a.r.POST("region/add_to_flight/:id", a.add_region_to_flight)
	a.r.GET("flights", a.get_flights)
	a.r.PUT("book", a.book)
	a.r.PUT("flight/status_change", a.flight_status_change)
	a.r.GET("flight_regions/:flight_id", a.flight_regions)
	a.r.PUT("flight/set_regions", a.set_flight_regions)

	a.r.Use(a.WithAuthCheck(role.Moderator, role.Admin)).PUT("region/delete_restore/:region_name", a.delete_restore_region)
	a.r.POST("region/add_image/:region_id", a.add_image)
	a.r.PUT("flight/delete/:flight_id", a.delete_flight)
	a.r.DELETE("flight_to_region/delete", a.delete_flight_to_region)
	a.r.PUT("flight/edit", a.edit_flight)
	a.r.DELETE("region/delete/:region_name", a.delete_region)
	a.r.PUT("region/edit", a.edit_region)
	a.r.POST("region/add", a.add_region)
	a.r.PUT("flight/set_allowed_hours", a.set_allowed_hours)

	a.r.Run(":80")

	log.Println("Server is down")
}

// @Summary Получить все регионы
// @Tags Регионы
// @Accept json
// @Produce json
// @Success 200 {} json
// @Param name_pattern query string false "Паттерн имени региона"
// @Param district query string false "Округ"
// @Param status query string false "Статус региона (Действует/Недействителен)"
// @Router /regions [get]

func (a *Application) get_regions(c *gin.Context) {
	var name_pattern = c.Query("name_pattern")
	var district = c.Query("district")
	var status = c.Query("status")

	regions, err := a.repo.GetRegions(name_pattern, district, status)
	if err != nil {
		c.Error(err)
		return
	}

	_userUUID, ok := c.Get("userUUID")

	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"regions": regions,
		})
		return
	}

	userUUID := _userUUID.(uuid.UUID)

	draft_flight, err := a.repo.GetDraftFlight(userUUID)

	if err != nil {
		log.Println(err)
		c.String(http.StatusInternalServerError, "Возникла ошибка при поиске заявки-черновика")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"regions":      regions,
		"draft_flight": draft_flight,
	})
}

// @Summary      Добавить регион в БД
// @Description  Создаёт новый регион с праметрами, описанными в json'е
// @Tags Регионы
// @Accept json
// @Produce      json
// @Param region body ds.Region true "Данные нового регионы"
// @Success      201  {object}  string "Регион был успешно создан"
// @Router       /region/add [put]
func (a *Application) add_region(c *gin.Context) {
	var region ds.Region

	if err := c.BindJSON(&region); err != nil || region.Name == "" || region.Status == "" {
		c.String(http.StatusBadRequest, "Невозможно распознать регион\n"+err.Error())
		return
	}

	if region.Status == "" {
		region.Status = "Черновик"
	}

	err := a.repo.CreateRegion(region)

	if err != nil {
		c.String(http.StatusNotFound, "Невозможно создать регион\n"+err.Error())
		return
	}

	c.String(http.StatusCreated, "Регион был успешно создан")

}

// @Summary      Получить регион
// @Description  Возвращает регион по имени
// @Tags         Регионы
// @Produce      json
// @Param region path string true "Имя региона"
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

// @Summary      Отредактировать регион
// @Description  Находит регион по имени и обновляет его поля
// @Tags         Регионы
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param region body ds.Region true "Новые данные изменяемого региона (должно быть имя региона или его id)"
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

// @Summary      Удалить регион
// @Description  Находит регион по имени и меняет его статус на "Недоступен"
// @Tags         Регионы
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param region_name path string true "Название региона"
// @Router       /region/delete/{region_name} [put]
func (a *Application) delete_region(c *gin.Context) {
	region_name := c.Param("region_name")

	if region_name == "" {
		c.String(http.StatusBadRequest, "Нужно предоставить имя региона")

		return
	}

	err := a.repo.LogicalDeleteRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Регион был успешно удалён")
}

// @Summary      Удаляет или восстанавливает регион
// @Description  Меняет статус региона с "Действует" на "Недоступен" и наобороь
// @Tags         Регионы
// @Produce      json
// @Success      200  {object}  string
// @Param region_name path string true "Имя региона"
// @Router       /region/delete_restore/{region_name} [get]
func (a *Application) delete_restore_region(c *gin.Context) {
	region_name := c.Param("region_name")

	if region_name == "" {
		c.String(http.StatusBadRequest, "Нужно предоставить название региона")
	}

	err := a.repo.DeleteRestoreRegion(region_name)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusFound, "Статус региона был успешно изменён")
}

func (a *Application) book(c *gin.Context) {
	var request_body ds.BookRequestBody

	if err := c.BindJSON(&request_body); err != nil {
		c.String(http.StatusBadGateway, "Не могу распознать json")
		return
	}

	_userUUID, ok := c.Get("userUUID")

	if !ok {
		c.String(http.StatusInternalServerError, "Вы сначала должны залогиниться")
		return
	}

	userUUID := _userUUID.(uuid.UUID)
	err := a.repo.Book(request_body, userUUID)

	if err != nil {
		c.Error(err)
		c.String(http.StatusNotFound, "Не могу забронировать регион")
		return
	}

	c.String(http.StatusCreated, "Бронирование прошло успешно!")
}

// @Summary      Получить заявки
// @Description  Возвращает список заявок
// @Tags         Заявки
// @Produce      json
// @Success      302  {object}  string
// @Param status query string false "Статус заявок"
// @Router       /flights [get]
func (a *Application) get_flights(c *gin.Context) {
	_roleNumber, _ := c.Get("role")
	_userUUID, _ := c.Get("userUUID")

	roleNumber := _roleNumber.(role.Role)
	userUUID := _userUUID.(uuid.UUID)

	status := c.Query("status")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	flights, err := a.repo.GetFlights(status, startDate, endDate, roleNumber, userUUID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, flights)
}

// @Summary      Получить заявку
// @Description  Возвращает заяввку с указанными параметрами
// @Tags         Заявки
// @Accept		 json
// @Produce      json
// @Success      302  {object}  string
// @Param status query string false "Статус заявки"
// @Param id query int false "ID заявки"
// @Router       /flight [get]
func (a *Application) get_flight(c *gin.Context) {
	status := c.Query("status")
	id, _ := strconv.ParseUint(c.Query("flight_id"), 10, 64)

	flight := &ds.Flight{
		Status: status,
		ID:     uint(id),
	}

	found_flight, err := a.repo.FindFlight(flight)

	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, found_flight)
}

// @Summary      Отредактировать заявку
// @Description  Находит заявку и обновляет её поля
// @Tags         Заявки
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param flight body ds.Flight false "Заявка"
// @Router       /flight/edit [put]
func (a *Application) edit_flight(c *gin.Context) {
	var requestBody ds.EditFlightRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.String(http.StatusBadRequest, "Передан плохой json")
		return
	}

	_userUUID, _ := c.Get("userUUID")
	userUUID := _userUUID.(uuid.UUID)

	var flight = ds.Flight{}
	flight.ArrivalDate = datatypes.Date(requestBody.ArrivalDate)
	flight.TakeoffDate = datatypes.Date(requestBody.TakeoffDate)
	flight.ID = uint(requestBody.FlightID)
	flight.Status = requestBody.Status

	err := a.repo.EditFlight(&flight, userUUID)

	if err != nil {
		c.Error(err)
		return
	}

	c.String(http.StatusCreated, "Заявка была успешно обновлена")
}

func (a *Application) flight_regions(c *gin.Context) { // нужно добавить проверку на авторизацию пользователя
	flight_id, err := strconv.Atoi(c.Param("flight_id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Не могу разобрать id полёта!")
		return
	}

	regions, err := a.repo.GetFlightRegions(flight_id)
	log.Println(regions)
	if err != nil {
		c.String(http.StatusInternalServerError, "Не получается достать регионы связанные с полётом!")
		return
	}

	c.JSON(http.StatusOK, regions)

}

func (a *Application) set_flight_regions(c *gin.Context) {
	var requestBody ds.SetFlightRegionsRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.String(http.StatusBadRequest, "Не получается распознать json запрос")
		return
	}

	err := a.repo.SetFlightRegions(requestBody.FlightID, requestBody.Regions)
	if err != nil {
		c.String(http.StatusInternalServerError, "Не получилось задать регионы для заявки\n"+err.Error())
	}

	c.String(http.StatusCreated, "Регионы заявки успешно заданы!")

}

// @Summary Изменить статус заявки
// @Description Получает id заявки и новый статус и производит необходимые обновления
// @Tags Заявки
// @Tags Заявки
// @Accept json
// @Produce json
// @Success 201 {object} string
// @Param request_body body ds.ChangeFlightStatusRequestBody true "Тело запроса"
// @Router /flight/status_change [put]
func (a *Application) flight_status_change(c *gin.Context) {
	var requestBody ds.ChangeFlightStatusRequestBody

	if err := c.BindJSON(&requestBody); err != nil {
		c.Error(err)
		return
	}

	_userUUID, _ := c.Get("userUUID")
	_userRole, _ := c.Get("role")

	userUUID := _userUUID.(uuid.UUID)
	userRole := _userRole.(role.Role)

	status, err := a.repo.GetFlightStatus(requestBody.ID)
	if err == nil {
		c.Error(err)
		return
	}

	if userRole == role.User && requestBody.Status == "Удалён" {

		if status == "Черновик" || status == "Сформирован" {
			err = a.repo.ChangeFlightStatusUser(requestBody.ID, requestBody.Status, userUUID)

			if err != nil {
				c.Error(err)
				return
			} else {
				c.String(http.StatusCreated, "Статус заявки был успешно обновлён")
			}
		}
	} else {
		err = a.repo.ChangeFlightStatus(requestBody.ID, requestBody.Status)

		if err != nil {
			c.Error(err)
			return
		}

		if userRole == role.Moderator && status == "Черновик" {
			err = a.repo.SetFlightModerator(requestBody.ID, userUUID)

			if err != nil {
				c.Error(err)
				return
			}

		}

		c.String(http.StatusCreated, "Статус заявки был успешно обновлён")
	}
}

// @Summary      Удалить заявку
// @Description  Меняет статус заявки на "Удалён"
// @Tags         Заявки
// @Summary      Удалить заявку
// @Description  Меняет статус заявки на "Удалён"
// @Tags         Заявки
// @Accept json
// @Produce      json
// @Success      302  {object}  string
// @Param flight_id path int true "id заявки"
// @Param flight_id path int true "id заявки"
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

// @Summary      Удаляет связь региона с заявкой
// @Tags         Заявки
// @Accept json
// @Produce      json
// @Success      201  {object}  string
// @Param request_body body ds.DeleteFlightToRegionRequestBody true "Тело запроса"
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

// @Summary Войти в систему
// @Description Возвращает jwt токен
// @Tags Аутентификация
// @Produce json
// @Accept json
// @Success 200 {object} loginResp
// @Param request_body body loginReq true "Тело запроса на вход"
// @Router /login [post]
func (a *Application) login(c *gin.Context) {
	req := &loginReq{}

	err := json.NewDecoder(c.Request.Body).Decode(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := a.repo.GetUserByLogin(req.Login)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if req.Login == user.Name && user.Pass == generateHashString(req.Password) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: time.Now().Add(3600000000000).Unix(),
				IssuedAt:  time.Now().Unix(),
				Issuer:    "dj1vs",
			},
			UserUUID: user.UUID,
			Scopes:   []string{}, // test data
			Role:     user.Role,
		})

		if token == nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("токен равен nil"))

			return
		}

		jwtToken := "test"

		strToken, err := token.SignedString([]byte(jwtToken))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("не получается просесть строку токена"))

			return
		}

		c.SetCookie("drones-api-token", "Bearer "+strToken, 3600000000000, "", "", true, true)

		c.JSON(http.StatusOK, loginResp{
			Login:       user.Name,
			Role:        int(user.Role),
			ExpiresIn:   3600000000000,
			AccessToken: strToken,
			TokenType:   "Bearer",
		})

		return
	}

	c.AbortWithStatus(http.StatusForbidden)
}

type registerReq struct {
	Login    string `json:"login"` // лучше назвать то же самое что login
	Password string `json:"password"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

// @Summary Зарегистрировать нового пользователя
// @Description Добавляет нового пользователя в БД
// @Tags Аутентификация
// @Produce json
// @Accept json
// @Success 200 {object} registerResp
// @Param request_body body registerReq true "Тело запроса"
// @Router /register [post]
func (a *Application) register(c *gin.Context) {
	req := &registerReq{}
	err := json.NewDecoder(c.Request.Body).Decode(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	if req.Password == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Password should not be empty"))
		return
	}
	if req.Login == "" {
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("Name should not be empty"))
	}

	err = a.repo.Register(&ds.User{
		UUID: uuid.New(),
		Role: role.User,
		Name: req.Login,
		Pass: generateHashString(req.Password),
	})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

// @Summary Выйти из системы
// @Details Деактивирует токен пользователя
// @Tags Аутентификация
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

type setAllowedHoursReq struct {
	flightId     int
	allowedHours string
}

func (a *Application) add_region_to_flight(c *gin.Context) {
	region_param := c.Param("id")

	region_id, err := strconv.Atoi(region_param)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	_userUUID, _ := c.Get("userUUID")
	userUUID := _userUUID.(uuid.UUID)

	draft, err := a.repo.GetDraftFlight(userUUID)
	if err != nil {
		c.String(http.StatusInternalServerError, "Не могу найти черновой полёт!")
	}

	region_to_draft := ds.FlightToRegion{}
	region_to_draft.FlightRefer = int(draft.ID)
	region_to_draft.RegionRefer = region_id

	err = a.repo.CreateFlightToRegion(region_to_draft)
	if err != nil {
		c.String(http.StatusInternalServerError, "Не могу связать район с полётом!")
	}

	c.String(http.StatusOK, "Район добавлен в черновой полёт!")

}

func (a *Application) set_allowed_hours(c *gin.Context) {
	req := &setAllowedHoursReq{}
	err := json.NewDecoder(c.Request.Body).Decode(req)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	flight := &ds.Flight{}
	flight.ID = uint(req.flightId)
	flight.AllowedHours = req.allowedHours

	_userUUID, _ := c.Get("userUUID")
	userUUID := _userUUID.(uuid.UUID)

	err = a.repo.EditFlight(flight, userUUID)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	c.String(http.StatusOK, "Разрешённые часы выставлены!")
}

func (a *Application) add_image(c *gin.Context) {
	region_id, err := strconv.Atoi(c.Param("region_id"))
	if err != nil {
		c.String(http.StatusBadRequest, "Не получается прочитать ID региона")
		log.Println("Не получается прочитать ID региона")
		return
	}

	image, header, err := c.Request.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, "Не получается распознать картинку")
		log.Println("Не получается распознать картинку")
		return
	}
	defer image.Close()

	minioClient, err := minio.New("127.0.0.1:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})

	if err != nil {
		c.String(http.StatusInternalServerError, "Не получается подключиться к minio")
		log.Println("Не получается подключиться к minio")
		return
	}

	objectName := header.Filename
	_, err = minioClient.PutObject(c.Request.Context(), "regionimages", objectName, image, header.Size, minio.PutObjectOptions{})

	if err != nil {
		c.String(http.StatusInternalServerError, "Не получилось загрузить картинку в minio")
		log.Println("Не получилось загрузить картинку в minio")
		return
	}

	err = a.repo.SetRegionImage(region_id, objectName)

	if err != nil {
		c.String(http.StatusInternalServerError, "Не получается обновить картинку региона")
		log.Println("Не получается обновить картинку региона")
		return
	}

	c.String(http.StatusCreated, "Картинка загружена!")

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
