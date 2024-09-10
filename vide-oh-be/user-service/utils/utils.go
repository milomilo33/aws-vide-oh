package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"user-service/auth"

	"github.com/gin-gonic/gin"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetTokenClaims(context *gin.Context) (err error, jwtClaims auth.JWTClaim) {
	tokenString := context.GetHeader("Authorization")
	err, jwtClaims = auth.ValidateToken(tokenString)
	return
}

type DBSecret struct {
	Host     string `json:"host"`
	Port     int16  `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type KeySecret struct {
	SecretKey string `json:"secretKey"`
}

func GetSecrets(secretName, keySecretName, region string) (*DBSecret, *KeySecret, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	dbInput := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	keyInput := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(keySecretName),
	}

	dbResult, err := client.GetSecretValue(context.TODO(), dbInput)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve secret: %v", err)
	}
	keyResult, err := client.GetSecretValue(context.TODO(), keyInput)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve secret: %v", err)
	}

	var dbSecret DBSecret
	err = json.Unmarshal([]byte(*dbResult.SecretString), &dbSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal secret: %v", err)
	}
	var keySecret KeySecret
	err = json.Unmarshal([]byte(*keyResult.SecretString), &keySecret)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal secret: %v", err)
	}

	return &dbSecret, &keySecret, nil
}
