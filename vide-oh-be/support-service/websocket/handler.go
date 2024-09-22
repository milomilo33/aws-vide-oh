package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"support-service/controllers"
	"support-service/models"
	"support-service/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func HandleConnect(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, tableNameConnections string, region string) (interface{}, error) {
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
	av, err := attributevalue.MarshalMap(m)
	if err != nil {
		log.Fatalln("Unable to marshal connection map", err.Error())
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableNameConnections),
		Item:      av,
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

func HandleDisconnect(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, tableNameConnections string, region string) (interface{}, error) {
	connectionID := req.RequestContext.ConnectionID

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

func HandleMessage(
	ctx context.Context,
	req events.APIGatewayWebsocketProxyRequest,
	tableNameConnections string,
	region string,
	websocketApiUrl string,
) (interface{}, error) {
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
	apigatewayClient := apigatewaymanagementapi.NewFromConfig(cfg, func(o *apigatewaymanagementapi.Options) {
		o.BaseEndpoint = aws.String(strings.Replace(websocketApiUrl, "wss", "https", 1))
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
