package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"

	"video-service/controllers"
	"video-service/database"
	"video-service/utils"

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
	// database.Migrate()

	// Initial Data
	// video := &models.Video{
	// 	Filename:    "someuniquefilename",
	// 	OwnerEmail:  "user@user.com",
	// 	Title:       "user's example video",
	// 	Description: "you'll see nothing special here",
	// }
	// database.Instance.Save(&video)

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

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func initRouter() *gin.Engine {
	router := gin.Default()
	router.Use(CORS())
	router.MaxMultipartMemory = 10 * 1024 * 1024
	api := router.Group("/api/videos")
	{
		// api.Static("/static", "./static")
		api.GET("/video-stream/:name", controllers.StreamVideo)
		api.GET("/report-video/:id", controllers.ReportVideo)
		api.GET("/search-videos", controllers.SearchVideos)

		// protected
		api.GET("/ping")
		api.GET("/all-reported-videos", controllers.GetAllReportedVideos)
		api.POST("/upload-video", controllers.UploadVideo)
		api.GET("/delete-video/:id", controllers.DeleteVideo)
	}
	return router
}
