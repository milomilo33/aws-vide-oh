package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"support-service/controllers"
	"support-service/database"
	"support-service/models"
	"support-service/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
				return HandleConnect(ctx, websocketRequest)
			case "$disconnect":
				return HandleDisconnect(ctx, websocketRequest)
			default:
				return HandleMessage(ctx, websocketRequest)
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

func HandleConnect(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (interface{}, error) {
	token := req.QueryStringParameters["token"]
	userEmail := req.QueryStringParameters["userEmail"]
	_, claims := utils.GetTokenClaimsFromTokenString(token)
	if claims.Role != "SupportUser" && !(claims.Role == "RegisteredUser" && claims.Email == userEmail) {
		return events.APIGatewayProxyResponse{
			StatusCode: 401,
			Body:       "unauthorized role",
		}, nil
	}

	connectionID := req.RequestContext.ConnectionID

	m := models.Connection{
		ID:           connectionID,
		ConnectionID: connectionID,
		UserEmail:    userEmail,
	}

	// Log the Connection struct before marshaling
	log.Printf("Connection struct: %+v\n", m)

	av, err := attributevalue.MarshalMap(m)
	if err != nil {
		log.Fatalln("Unable to marshal connection map", err.Error())
	}

	// Log the marshaled attributes
	log.Printf("Marshaled item: %+v\n", av)

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableNameConnections),
		Item:      av,
	}

	// Log the marshaled attributes
	log.Printf("Input: %+v\n", input)

	cfg, err := utils.GetSession(region)
	if err != nil {
		log.Fatalln("Unable to get AWS session")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Unable to get AWS session",
		}, nil
	}
	db := dynamodb.NewFromConfig(cfg)

	_, err = db.PutItem(ctx, input)
	if err != nil {
		log.Fatal("INSERT ERROR", err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Unable to insert connection",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func HandleDisconnect(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (interface{}, error) {
	connectionID := req.RequestContext.ConnectionID

	// Create DeleteItemInput for DynamoDB in AWS SDK V2
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableNameConnections),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: connectionID},
		},
	}

	cfg, err := utils.GetSession(region)
	if err != nil {
		log.Fatalln("Unable to get AWS session")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Unable to get AWS session",
		}, nil
	}
	db := dynamodb.NewFromConfig(cfg)

	_, err = db.DeleteItem(ctx, input)
	if err != nil {
		log.Fatalln("Unable to remove connection from DynamoDB", err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Unable to remove connection",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func HandleMessage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (interface{}, error) {
	connectionID := req.RequestContext.ConnectionID

	// Convert message to JSON and parse JWT claims
	socketMessage := models.SocketMessage{}
	if err := json.NewDecoder(strings.NewReader(req.Body)).Decode(&socketMessage); err != nil {
		log.Println("Unable to decode body", err.Error())
	}
	_, claims := utils.GetTokenClaimsFromTokenString(socketMessage.Token)

	cfg, err := utils.GetSession(region)
	if err != nil {
		log.Fatalln("Unable to get AWS session")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Unable to get AWS session",
		}, nil
	}
	db := dynamodb.NewFromConfig(cfg)

	// Get connection (with user email) for connection ID
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(tableNameConnections),
		ExpressionAttributeNames: map[string]string{
			"#id": "id",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: connectionID},
		},
		KeyConditionExpression: aws.String("#id = :id"),
	}
	result, err := db.Query(ctx, queryInput)
	if err != nil {
		log.Println("Unable to find connection ID", err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Unable to find connection ID",
		}, nil
	}
	connections := make([]models.Connection, result.Count)
	attributevalue.UnmarshalListOfMaps(result.Items, &connections)
	if len(connections) > 1 {
		log.Println("Found duplicate connection ID")
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Found duplicate connection ID",
		}, nil
	}
	connection := connections[0]

	// Add new message to DB
	message, err := controllers.AddMessage(socketMessage.Message, connection.UserEmail, claims)
	if err != nil {
		fmt.Println("error: ", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error adding message to DB",
		}, nil
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Println("error: ", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error marshalling message to JSON",
		}, nil
	}

	// Get all connections for user email
	params := &dynamodb.ScanInput{
		TableName:        aws.String(tableNameConnections),
		FilterExpression: aws.String("userEmail = :userEmail"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userEmail": &types.AttributeValueMemberS{Value: connection.UserEmail},
		},
	}
	scanResult, err := db.Scan(ctx, params)
	if err != nil {
		log.Println("Failed to scan user email", err.Error())
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Failed to scan user email",
		}, nil
	}
	connectionsWithUserEmail := make([]models.Connection, scanResult.Count)
	attributevalue.UnmarshalListOfMaps(scanResult.Items, &connectionsWithUserEmail)

	// Send message to relevant connections
	// logger := log.New(os.Stdout, "", log.LstdFlags)
	// loggingImpl := logging.LoggerFunc(func(classification logging.Classification, format string, v ...interface{}) {
	// 	logger.Printf("[%v] %s", classification, fmt.Sprintf(format, v...))
	// })
	// cfgx, err := config.LoadDefaultConfig(context.TODO(),
	// 	config.WithLogger(loggingImpl),
	// 	config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody),
	// 	config.WithRegion(region),
	// )
	x := strings.Replace(websocketApiUrl, "wss", "https", 1)
	fmt.Println(x)
	apigatewayClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(x)
	})
	for _, connectionWithUserEmail := range connectionsWithUserEmail {
		input := &apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionWithUserEmail.ConnectionID),
			Data:         messageJSON,
		}
		_, err = apigatewayClient.PostToConnection(context.TODO(), input)
		if err != nil {
			log.Println("ERROR", err.Error())
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}
