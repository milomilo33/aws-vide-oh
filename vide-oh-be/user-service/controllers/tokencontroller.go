package controllers

import (
	"fmt"
	"net/http"
	"user-service/auth"
	"user-service/database"
	"user-service/models"

	"github.com/gin-gonic/gin"
)

type TokenRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(context *gin.Context) {
	var request TokenRequest
	var user models.User
	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		fmt.Println("xd")
		context.Abort()
		return
	}
	fmt.Println(request)

	record := database.Instance.Where("email = ?", request.Email).First(&user)
	if record.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": record.Error.Error()})
		fmt.Println("xd2")
		context.Abort()
		return
	}

	credentialError := user.CheckPassword(request.Password)
	if credentialError != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		fmt.Println("xd3")
		context.Abort()
		return
	}

	tokenString, err := auth.GenerateJWT(user.Email, user.Role.String())
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println("xd4")
		context.Abort()
		return
	}
	context.JSON(http.StatusOK, gin.H{"token": tokenString})
}
