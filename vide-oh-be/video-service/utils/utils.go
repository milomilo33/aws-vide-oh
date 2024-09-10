package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
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

type DBSecret struct {
	Host     string `json:"host"`
	Port     int16  `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

func GetSecret(secretName, region string) (*DBSecret, error) {
	// Load the default AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	// Create a Secrets Manager client
	client := secretsmanager.NewFromConfig(cfg)

	// Get the secret value
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := client.GetSecretValue(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secret: %v", err)
	}

	// Parse the secret string
	var secret DBSecret
	err = json.Unmarshal([]byte(*result.SecretString), &secret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %v", err)
	}

	return &secret, nil
}
