package middleware

import (
	"errors"
	"user-service/auth"
	"user-service/database"
	"user-service/models"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		tokenString := context.GetHeader("Authorization")
		if tokenString == "" {
			context.JSON(401, gin.H{"error": "request does not contain an access token"})
			context.Abort()
			return
		}
		err, claims := auth.ValidateToken(tokenString)
		if err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}

		// auth invalid if user blocked
		var user models.User

		if err := database.Instance.Where("email = ?", claims.Email).First(&user).Error; err != nil {
			context.JSON(401, gin.H{"error": err.Error()})
			context.Abort()
			return
		}
		if user.Blocked {
			context.JSON(401, gin.H{"error": "user blocked"})
			context.Abort()
			return
		}

		context.Next()
	}
}

func ValidateTokenForLambdaAuthorizer(token string) (err error, jwtClaims auth.JWTClaim) {
	err, claims := auth.ValidateToken(token)
	if err != nil {
		return
	}

	// auth invalid if user blocked
	var user models.User
	if err = database.Instance.Where("email = ?", claims.Email).First(&user).Error; err != nil {
		return
	}
	if user.Blocked {
		err = errors.New("user account is blocked")
		return
	}

	jwtClaims = claims

	return
}
