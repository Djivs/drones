package app

import (
	"drones/internal/app/ds"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const jwtPrefix = "Bearer "

func (a *Application) WithAuthCheck(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")

	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		c.AbortWithStatus(http.StatusForbidden)

		return
	}

	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test"), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		log.Println(err)

		return
	}

}
