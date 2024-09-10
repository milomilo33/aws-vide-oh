package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"user-service/controllers"
	"user-service/database"
	"user-service/middleware"
	"user-service/models"
	"user-service/utils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

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

	// Initialize Database
	database.Connect(connectionString)
	database.Migrate()

	// Initial Data
	user := &models.User{
		Name:     "Admin Adminsky",
		Email:    "admin@admin.com",
		Password: "123",
		Role:     models.Administrator,
		Blocked:  false,
	}
	user.HashPassword(user.Password)
	database.Instance.Save(&user)

	user2 := &models.User{
		Name:     "User Usersky",
		Email:    "user@user.com",
		Password: "123",
		Role:     models.RegisteredUser,
		Blocked:  false,
	}
	user2.HashPassword(user2.Password)
	database.Instance.Save(&user2)

	user3 := &models.User{
		Name:     "User2 Usersky2",
		Email:    "user2@user2.com",
		Password: "123",
		Role:     models.RegisteredUser,
		Blocked:  false,
	}
	user3.HashPassword(user3.Password)
	database.Instance.Save(&user3)

	supportUser := &models.User{
		Name:     "Tech Support",
		Email:    "tech@support.com",
		Password: "123",
		Role:     models.SupportUser,
		Blocked:  false,
	}
	supportUser.HashPassword(supportUser.Password)
	database.Instance.Save(&supportUser)

	// Start the Lambda handler
	lambda.Start(Handler)

	fmt.Println("User service lambda started!")
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/users")
	{
		api.POST("/login", controllers.Login)
		api.POST("/register", controllers.RegisterUser)
		api.GET("/ping", controllers.Ping)
		secured := api.Group("/secured").Use(middleware.Auth())
		{
			secured.GET("/ping", controllers.Ping)
			secured.GET("/user/all-registered", controllers.GetAllRegisteredUsers) // only admin
			secured.GET("/block/:email", controllers.BlockUser)                    // only admin
			secured.GET("/user/:id", controllers.GetUserById)
			secured.GET("/user/current", controllers.GetCurrentUser)
			secured.GET("/user/change-name", controllers.ChangeName)
		}
	}
	return router
}
