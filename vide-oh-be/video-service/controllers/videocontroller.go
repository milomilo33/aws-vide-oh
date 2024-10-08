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

func GetAllReportedVideos(c *gin.Context) {
	_, claims := utils.GetTokenClaims(c)
	if claims.Role != "Administrator" {
		c.JSON(401, gin.H{"error": "unauthorized role"})
		c.Abort()
		return
	}

	var videos []models.Video
	database.Instance.Where("reported = ?", true).Find(&videos)

	var videoSearchResults []models.VideoSearchResultDTO

	// Generate pre-signed URLs for thumbnails and map the original videos to DTO
	for _, video := range videos {
		req := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(video.Filename + ".png"),
		}
		presignedURL, err := presignClient.PresignGetObject(context.TODO(), req, func(p *s3.PresignOptions) {
			p.Expires = 15 * time.Minute
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pre-signed URL"})
			c.Abort()
			return
		}

		videoSearchResults = append(videoSearchResults, toVideoSearchResultDTO(video, presignedURL.URL))
	}

	c.JSON(http.StatusOK, videoSearchResults)
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

func toVideoSearchResultDTO(video models.Video, thumbnailURL string) models.VideoSearchResultDTO {
	return models.VideoSearchResultDTO{
		ID:           video.ID,
		Title:        video.Title,
		Filename:     video.Filename,
		Description:  video.Description,
		OwnerEmail:   video.OwnerEmail,
		Reported:     video.Reported,
		ThumbnailURL: thumbnailURL,
	}
}

func SearchVideos(c *gin.Context) {
	var videos []models.Video
	searchQuery := c.Query("query")

	if searchQuery == "" {
		database.Instance.Find(&videos)
	} else {
		searchQuery = "%" + strings.ToLower(searchQuery) + "%"
		database.Instance.Where("lower(title) LIKE ?", searchQuery).
			Or("lower(description) LIKE ?", searchQuery).
			Or("owner_email LIKE ?", searchQuery).
			Find(&videos)
	}

	var videoSearchResults []models.VideoSearchResultDTO

	// Generate pre-signed URLs for thumbnails and map the original videos to DTO
	for _, video := range videos {
		req := &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(video.Filename + ".png"),
		}
		presignedURL, err := presignClient.PresignGetObject(context.TODO(), req, func(p *s3.PresignOptions) {
			p.Expires = 15 * time.Minute
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate pre-signed URL"})
			c.Abort()
			return
		}

		videoSearchResults = append(videoSearchResults, toVideoSearchResultDTO(video, presignedURL.URL))
	}

	c.JSON(http.StatusOK, videoSearchResults)
}
