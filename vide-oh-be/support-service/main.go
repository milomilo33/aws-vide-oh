package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"support-service/controllers"
	"support-service/database"
	"support-service/utils"
	"support-service/websocket"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

var ginLambda *ginadapter.GinLambda

var region string
var tableNameConnections string
var websocketApiUrl string

func init() {
	// Initialize Router
	router := initRouter()

	// Create the Lambda handler
	ginLambda = ginadapter.New(router)
}

func main() {
	// Load env vars
	secretName := os.Getenv("DB_SECRET_NAME")
	if secretName == "" {
		log.Fatal("DB_SECRET_NAME environment variable is not set")
	}
	tableNameConnections = os.Getenv("TABLE_NAME_CONNECTIONS")
	if tableNameConnections == "" {
		log.Fatal("TABLE_NAME_CONNECTIONS environment variable is not set")
	}
	region = os.Getenv("REGION")
	if secretName == "" {
		log.Fatal("REGION environment variable is not set")
	}
	websocketApiUrl = os.Getenv("WEBSOCKET_API_URL")
	if secretName == "" {
		log.Fatal("WEBSOCKET_API_URL environment variable is not set")
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

	// Initialize Database
	database.Connect(connectionString)
	database.Migrate()

	// Start the Lambda handler
	lambda.Start(Handler)
}

func CORS() gin.HandlerFunc {
	// TO allow CORS
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// Handler handles both WebSocket and HTTP API Gateway requests
func Handler(ctx context.Context, req interface{}) (interface{}, error) {

	// Try to cast to HTTP API Gateway (REST) request
	if request, ok := req.(map[string]interface{}); ok {
		// First, check if it's a REST API Gateway request
		if method, found := request["httpMethod"]; found && method != nil {
			fmt.Println("Handling HTTP API Gateway request")

			// Convert the raw map to APIGatewayProxyRequest
			var proxyRequest events.APIGatewayProxyRequest
			err := mapstructure.Decode(request, &proxyRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to decode request to APIGatewayProxyRequest: %v", err)
			}

			// Call your REST handler here (e.g., Gin Lambda Proxy)
			return ginLambda.ProxyWithContext(ctx, proxyRequest)
		}

		// If it's not an HTTP request, check if it's a WebSocket request
		if routeKey, found := request["requestContext"].(map[string]interface{})["routeKey"]; found && routeKey != nil {
			fmt.Println("Handling WebSocket API Gateway request")

			// Convert the raw map to APIGatewayWebsocketProxyRequest
			var websocketRequest events.APIGatewayWebsocketProxyRequest
			err := mapstructure.Decode(request, &websocketRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to decode request to APIGatewayWebsocketProxyRequest: %v", err)
			}

			// Handle WebSocket connections
			switch websocketRequest.RequestContext.RouteKey {
			case "$connect":
				return websocket.HandleConnect(ctx, websocketRequest, tableNameConnections, region)
			case "$disconnect":
				return websocket.HandleDisconnect(ctx, websocketRequest, tableNameConnections, region)
			default:
				return websocket.HandleMessage(ctx, websocketRequest, tableNameConnections, region, websocketApiUrl)
			}
		}
	}

	// If none of the above types matched, return an error
	fmt.Println("Unknown request type")
	return nil, fmt.Errorf("unknown request type")
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(CORS())
	api := router.Group("/api/messages")
	{
		api.GET("/:email/all", controllers.GetAllMessagesForUser)
		api.GET("/user-emails", controllers.GetAllUserEmailsWithMessages)
	}
	return router
}
