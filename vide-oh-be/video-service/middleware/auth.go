package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(context *gin.Context) {
		requestURL := os.Getenv("USER_SECURED_API_URL")
		if requestURL == "" {
			log.Fatal("USER_SECURED_API_URL environment variable is not set")
		}

		req, err := http.NewRequest(http.MethodGet, requestURL, nil)
		if err != nil {
			fmt.Printf("auth client: could not create request: %s\n", err)
			context.JSON(401, gin.H{"error": "auth client: could not create request: " + err.Error()})
			context.Abort()
			return
		}
		req.Header.Add("Authorization", context.GetHeader("Authorization"))
		req.Header.Add("x-api-key", context.GetHeader("x-api-key"))
		req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:60.0) Gecko/20100101 Firefox/60.0")
		// req.Header.Add("Accept", context.GetHeader("Accept"))
		// req.Header.Add("Accept-Encoding", context.GetHeader("Accept-Encoding"))
		// req.Header.Add("Connection", context.GetHeader("Connection"))

		// Print request details
		fmt.Printf("Request URL: %s\n", requestURL)
		fmt.Printf("Request Headers:\n")
		for key, values := range req.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("auth client: error making http request: %s\n", err)
			context.JSON(401, gin.H{"error": "auth client: error making http request: " + err.Error()})
			context.Abort()
			return
		}

		// Print the status and response body
		fmt.Printf("Response status: %s\n", res.Status)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("auth client: error reading response body: %s\n", err)
		} else {
			fmt.Printf("Response body: %s\n", string(body))
		}
		fmt.Printf("Request URL from res: %s\n", res.Request.URL)
		fmt.Printf("Request Headers from res:\n")
		for key, values := range res.Request.Header {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}

		if res.StatusCode != http.StatusOK {
			context.JSON(401, gin.H{"error": "token is not authorized"})
			context.Abort()
			return
		}

		context.Next()
	}
}
