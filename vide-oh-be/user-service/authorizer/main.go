package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"user-service/auth"
	"user-service/database"
	"user-service/middleware"
	"user-service/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayV2CustomAuthorizerV1Request) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := request.Headers["Authorization"]
	if token == "" {
		return generateUnauthorizedResponse("Authorization token is missing"), nil
	}
	fmt.Println("hi")
	// Validate token
	err, claims := middleware.ValidateTokenForLambdaAuthorizer(token)
	if err != nil {
		fmt.Println(err)
		return generateForbiddenResponse(fmt.Sprintf("unauthorized: %v", err)), nil
	}

	// Create policy allowing access to the resource
	policy := generatePolicy(claims.Email, "Allow", request.MethodArn)

	return policy, nil
}

// Function to generate a 403 Forbidden response
func generateForbiddenResponse(message string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Deny",
					Resource: []string{"*"},
				},
			},
		},
		Context: map[string]interface{}{
			"message": message,
		},
	}
}

// Function to generate a 401 Unauthorized response
func generateUnauthorizedResponse(message string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "user",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Deny",
					Resource: []string{"*"},
				},
			},
		},
		Context: map[string]interface{}{
			"message":    message,
			"statusCode": 401,
		},
	}
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
	dbSecretName := os.Getenv("DB_SECRET_NAME")
	if dbSecretName == "" {
		log.Fatal("DB_SECRET_NAME environment variable is not set")
	}
	region := os.Getenv("REGION")
	if region == "" {
		log.Fatal("REGION environment variable is not set")
	}
	keySecretName := os.Getenv("KEY_SECRET_NAME")
	if keySecretName == "" {
		log.Fatal("KEY_SECRET_NAME environment variable is not set")
	}

	// Fetch secrets from SM
	dbSecret, keySecret, err := utils.GetSecrets(dbSecretName, keySecretName, region)
	if err != nil {
		log.Fatalf("Failed to retrieve secret: %v", err)
	}
	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		dbSecret.Username,
		url.QueryEscape(dbSecret.Password),
		dbSecret.Host,
		dbSecret.Port,
		dbSecret.DBName,
	)
	auth.SetJwtKey(keySecret.SecretKey)

	database.Connect(connectionString)

	lambda.Start(handler)
}
