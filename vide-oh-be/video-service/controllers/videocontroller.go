package controllers

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"video-service/database"
	"video-service/models"
	"video-service/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
)

const CHUNK_SIZE int64 = 1024 * 1024

var (
	s3Client      *s3.Client
	bucketName    = "vide-oh-videos"
	presignClient *s3.PresignClient
)

// S3 client
func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)
}

// func StreamVideo(context *gin.Context) {
// 	rangeHeader := context.Request.Header["Range"][0]
// 	ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
// 	start, err := strconv.ParseInt(ranges[0], 10, 64)
// 	if err != nil {
// 		context.Status(http.StatusNotAcceptable)
// 		context.Abort()
// 		return
// 	}
// 	// var end int64
// 	// if len(ranges) == 2 && len(ranges[1]) > 0 {
// 	// 	end, err = strconv.ParseInt(ranges[1], 10, 64)
// 	// 	if err != nil {
// 	// 		context.Status(http.StatusNotAcceptable)
// 	// 		context.Abort()
// 	// 		return
// 	// 	}
// 	// } else {
// 	// 	end = start + CHUNK_SIZE
// 	// }

// 	// binary read
// 	name := context.Param("name")
// 	file, err := os.Open("static/" + name + ".mp4")

// 	if err != nil {
// 		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		context.Abort()
// 		return
// 	}
// 	defer file.Close()

// 	stats, statsErr := file.Stat()
// 	if statsErr != nil {
// 		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
// 		context.Abort()
// 		return
// 	}

// 	var totalFileSize int64 = stats.Size()
// 	end := start + CHUNK_SIZE
// 	if end > totalFileSize-1 {
// 		end = totalFileSize - 1
// 	}

// 	contentLength := end - start + 1
// 	data := make([]byte, contentLength)
// 	file.Seek(start, 0)
// 	bytesRead, _ := file.Read(data)
// 	fmt.Println("Bytes read: " + strconv.Itoa(bytesRead))

// 	context.Writer.Header().Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(totalFileSize, 10))
// 	context.Writer.Header().Add("Accept-Ranges", "bytes")
// 	context.Writer.Header().Add("Content-Length", strconv.FormatInt(contentLength, 10))
// 	context.Data(206, "video/mp4", data)
// }

func StreamVideo(c *gin.Context) {
	videoFilename := c.Param("name") + ".mp4"

	// Generate a pre-signed URL for the video object
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(videoFilename),
	}

	presignedURL, err := presignClient.PresignGetObject(context.TODO(), req, func(p *s3.PresignOptions) {
		p.Expires = 15 * time.Minute // URL valid for 15 minutes
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pre-signed URL"})
		c.Abort()
		return
	}

	// Return the pre-signed URL
	c.JSON(http.StatusOK, gin.H{"url": presignedURL.URL})
}

// func StreamVideo(c *gin.Context) {
// 	videoFilename := c.Param("name") + ".mp4"
// 	rangeHeader := c.Request.Header.Get("Range")

// 	if rangeHeader == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Range header required"})
// 		c.Abort()
// 		return
// 	}

// 	videoObject, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(videoFilename),
// 		Range:  aws.String(rangeHeader),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video from S3"})
// 		c.Abort()
// 		return
// 	}
// 	defer videoObject.Body.Close()

// 	contentRange := *videoObject.ContentRange
// 	contentLength := *videoObject.ContentLength

// 	c.Writer.Header().Set("Content-Range", contentRange)
// 	c.Writer.Header().Set("Accept-Ranges", "bytes")
// 	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
// 	c.Writer.Header().Set("Content-Type", "video/mp4")

// 	_, err = io.Copy(c.Writer, videoObject.Body)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stream video content"})
// 		c.Abort()
// 		return
// 	}

// 	c.Status(http.StatusPartialContent)
// }

// func StreamVideo(c *gin.Context) {
// 	// Get the filename from the URL
// 	videoFilename := c.Param("name") + ".mp4"

// 	// Get the Range header
// 	rangeHeader := c.Request.Header.Get("Range")
// 	if rangeHeader == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Range header required"})
// 		c.Abort()
// 		return
// 	}
// 	fmt.Println("Range header: ")
// 	fmt.Println(rangeHeader)
// 	// // Parse range
// 	// ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
// 	// start, err := strconv.ParseInt(ranges[0], 10, 64)
// 	// if err != nil {
// 	// 	c.Status(http.StatusNotAcceptable)
// 	// 	c.Abort()
// 	// 	return
// 	// }

// 	// Download video from S3
// 	videoObject, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(videoFilename),
// 		Range:  aws.String(rangeHeader), // Fetch only the range of bytes requested
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video from S3"})
// 		c.Abort()
// 		return
// 	}
// 	defer videoObject.Body.Close()

// 	// Read video content
// 	contentRange := *videoObject.ContentRange
// 	contentLength := *videoObject.ContentLength
// 	fmt.Println("range: ")
// 	fmt.Println(contentRange)
// 	fmt.Println("length: ")
// 	fmt.Println(contentLength)
// 	buffer := make([]byte, contentLength)
// 	videoObject.Body.Read(buffer)

// 	// Set headers for range-based response
// 	c.Writer.Header().Set("Content-Range", contentRange)
// 	c.Writer.Header().Set("Accept-Ranges", "bytes")
// 	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))

// 	// Return partial content status and video data
// 	c.Data(http.StatusPartialContent, "video/mp4", buffer)
// }

// func StreamVideo(c *gin.Context) {
// 	// Get the filename from the URL
// 	videoFilename := c.Param("name") + ".mp4"

// 	// Get the Range header
// 	rangeHeader := c.Request.Header.Get("Range")
// 	if rangeHeader == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Range header required"})
// 		c.Abort()
// 		return
// 	}

// 	// Parse range
// 	ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
// 	start, err := strconv.ParseInt(ranges[0], 10, 64)
// 	if err != nil {
// 		c.Status(http.StatusNotAcceptable)
// 		c.Abort()
// 		return
// 	}

// 	var end int64
// 	if len(ranges) > 1 && ranges[1] != "" {
// 		end, err = strconv.ParseInt(ranges[1], 10, 64)
// 		if err != nil {
// 			c.Status(http.StatusNotAcceptable)
// 			c.Abort()
// 			return
// 		}
// 	} else {
// 		end = start + 1024*1024 - 1 // Default to 1MB range if end is not specified
// 	}
// 	fmt.Println("Start: ")
// 	fmt.Println(start)
// 	fmt.Println("End: ")
// 	fmt.Println(end)
// 	// Download video from S3
// 	videoObject, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(videoFilename),
// 		Range:  aws.String(fmt.Sprintf("bytes=%d-%d", start, end)),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video from S3"})
// 		c.Abort()
// 		return
// 	}
// 	defer videoObject.Body.Close()

// 	// Read video content
// 	contentLength := end - start + 1
// 	buffer := make([]byte, contentLength)
// 	bytesRead, err := io.ReadFull(videoObject.Body, buffer)
// 	if err != nil && err != io.EOF {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read video content"})
// 		c.Abort()
// 		return
// 	}

// 	if int64(bytesRead) != contentLength {
// 		// Debugging output
// 		fmt.Printf("Expected content length: %d\n", contentLength)
// 		fmt.Printf("Actual content length read: %d\n", bytesRead)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected content length read"})
// 		c.Abort()
// 		return
// 	}

// 	// Set headers for range-based response
// 	c.Writer.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, *videoObject.ContentLength))
// 	c.Writer.Header().Set("Accept-Ranges", "bytes")
// 	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", bytesRead))
// 	c.Writer.Header().Set("Content-Type", "video/mp4")

// 	// Return partial content status and video data
// 	c.Data(http.StatusPartialContent, "video/mp4", buffer[:bytesRead])
// }

// func StreamVideo(c *gin.Context) {
// 	// Get the filename from the URL
// 	videoFilename := c.Param("name") + ".mp4"

// 	// Download the full video from S3
// 	videoObject, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(videoFilename),
// 	})
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video from S3"})
// 		c.Abort()
// 		return
// 	}
// 	defer videoObject.Body.Close()

// 	// Set headers for the full video response
// 	c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", *videoObject.ContentLength))
// 	c.Writer.Header().Set("Content-Type", "video/mp4")
// 	c.Writer.Header().Set("isBase64Encoded", "true")

// 	// Stream the video content directly from S3 to the response
// 	_, err = io.Copy(c.Writer, videoObject.Body)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to stream video content"})
// 		c.Abort()
// 		return
// 	}

// 	// Return success status
// 	c.Status(http.StatusOK)
// }

func ReportVideo(context *gin.Context) {
	videoId := context.Param("id")
	var video models.Video

	if err := database.Instance.First(&video, videoId).Error; err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	video.Reported = true

	database.Instance.Save(&video)

	context.Status(http.StatusOK)
}

func GetAllReportedVideos(context *gin.Context) {
	_, claims := utils.GetTokenClaims(context)
	if claims.Role != "Administrator" {
		context.JSON(401, gin.H{"error": "unauthorized role"})
		context.Abort()
		return
	}

	var videos []models.Video
	database.Instance.Where("reported = ?", true).Find(&videos)

	context.JSON(http.StatusOK, videos)
}

func UploadVideo(c *gin.Context) {
	_, claims := utils.GetTokenClaims(c)
	if claims.Role != "RegisteredUser" {
		c.JSON(401, gin.H{"error": "unauthorized role"})
		c.Abort()
		return
	}

	// single file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to retrieve file"})
		return
	}

	if filepath.Ext(file.Filename) != ".mp4" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid extension"})
		c.Abort()
		return
	}

	// Generate random filename
	rand.Seed(time.Now().UnixNano())
	rndNum := rand.Intn(math.MaxInt32-0) + 0
	filenameNoExt := strconv.Itoa(rndNum)
	videoFilename := filenameNoExt + ".mp4"

	// Save video file to temporary Lambda storage
	err = c.SaveUploadedFile(file, "/tmp/"+videoFilename)
	if err != nil {
		fmt.Println(err.Error())
	}

	// Generate thumbnail from the video file
	err = generateVideoThumbnailFromFile(filenameNoExt)
	if err != nil {
		log.Println("Failed to generate thumbnail:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate thumbnail"})
		c.Abort()
		return
	}

	// Open the video file to upload to S3
	src, err := file.Open()
	if err != nil {
		log.Println("Failed to reopen video file for upload:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open video file"})
		return
	}
	defer src.Close()

	// Upload video file to S3
	err = uploadToS3(src, videoFilename, "video/mp4")
	if err != nil {
		log.Println("Failed to upload video to S3:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload video to S3"})
		return
	}

	// Save video metadata to DB
	video := &models.Video{
		Title:       c.Query("title"),
		Description: c.Query("description"),
		OwnerEmail:  claims.Email,
		Filename:    filenameNoExt,
	}
	database.Instance.Save(&video)

	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", videoFilename))
}

func uploadToS3(file multipart.File, filename, contentType string) error {
	// Read file content
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	// Upload to S3
	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	return err
}

func generateVideoThumbnailFromFile(filenameNoExt string) error {
	// Use /tmp for temporary file storage in Lambda
	videoPath := fmt.Sprintf("/tmp/%s.mp4", filenameNoExt)
	outputFilePath := fmt.Sprintf("/tmp/%s.png", filenameNoExt)

	// Generate the thumbnail using ffmpeg
	cmd := fmt.Sprintf(`./ffmpeg -i "%s" -an -q 0 -vf "scale='if(gt(iw\,ih)\,-1\,200)':'if(gt(iw\,ih)\,200\,-1)',crop=200:200:exact=1" -vframes 1 "%s"`, videoPath, outputFilePath)
	shellName := "bash"
	ffCmd := exec.Command(shellName, "-c", cmd)
	output, err := ffCmd.CombinedOutput()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + string(output))
		return err
	}
	log.Println(string(output))

	// Upload thumbnail to S3
	thumbnailFile, err := os.Open(outputFilePath)
	if err != nil {
		log.Println("Failed to open thumbnail file:", err)
		return err
	}
	defer thumbnailFile.Close()
	err = uploadToS3(thumbnailFile, fmt.Sprintf("%s.png", filenameNoExt), "image/png")
	if err != nil {
		return err
	}

	// Clean up the temporary files
	if err := os.Remove(videoPath); err != nil {
		log.Println("Failed to delete video file:", err)
		return err
	}
	if err := os.Remove(outputFilePath); err != nil {
		log.Println("Failed to delete thumbnail file:", err)
		return err
	}

	return nil
}

func DeleteVideo(context *gin.Context) {
	videoId := context.Param("id")
	var video models.Video

	if err := database.Instance.First(&video, videoId).Error; err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	_, claims := utils.GetTokenClaims(context)
	if claims.Role == "RegisteredUser" && claims.Email != video.OwnerEmail {
		context.JSON(401, gin.H{"error": "you are not authorized to delete this video"})
		context.Abort()
		return
	}

	database.Instance.Delete(&models.Video{}, videoId)

	context.Status(http.StatusOK)
}

func SearchVideos(context *gin.Context) {
	var videos []models.Video
	searchQuery := context.Query("query")
	if searchQuery == "" {
		database.Instance.Find(&videos)
		context.JSON(http.StatusOK, videos)
		context.Abort()
		return
	}

	searchQuery = "%" + strings.ToLower(searchQuery) + "%"
	database.Instance.Where("lower(title) LIKE ?", searchQuery).Or("lower(description) LIKE ?", searchQuery).Or("owner_email LIKE ?", searchQuery).Find(&videos)
	context.JSON(http.StatusOK, videos)
}
