package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"user-service/database"
	"user-service/middleware"
	"user-service/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayV2CustomAuthorizerV1Request) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := request.Headers["Authorization"]
	err, claims := middleware.ValidateTokenForLambdaAuthorizer(token)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("unauthorized: %v", err)
	}

	// Create policy allowing access to the resource
	policy := generatePolicy(claims.Email, "Allow", request.MethodArn)

	return policy, nil
}

func generatePolicy(principalId, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalId,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		},
	}
}

func main() {
	// Load env vars
	secretName := os.Getenv("DB_SECRET_NAME")
	if secretName == "" {
		log.Fatal("DB_SECRET_NAME environment variable is not set")
	}
	region := os.Getenv("REGION")
	if secretName == "" {
		log.Fatal("REGION environment variable is not set")
	}

	// Read AWS secret DB connection info
	secret, err := utils.GetSecret(secretName, region)
	if err != nil {
		log.Fatalf("Failed to retrieve secret: %v", err)
	}
	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		secret.Username,
		url.QueryEscape(secret.Password),
		secret.Host,
		secret.Port,
		secret.DBName,
	)

	database.Connect(connectionString)
	database.Migrate()

	lambda.Start(handler)
}
