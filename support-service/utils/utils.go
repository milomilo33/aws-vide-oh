package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/dgrijalva/jwt-go"
)

type JWTClaim struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

func GetTokenClaims(context *gin.Context) (err error, jwtClaims JWTClaim) {
	tokenString := context.GetHeader("Authorization")
	parts := strings.Split(tokenString, ".")
	bytes, _ := base64.RawURLEncoding.DecodeString(parts[1])
	err = json.Unmarshal(bytes, &jwtClaims)
	if err != nil {
		fmt.Println("error: ", err)
	}
	return
}
