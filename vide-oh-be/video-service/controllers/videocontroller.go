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

var s3Client *s3.Client
var bucketName = "vide-oh-videos"

// S3 client
func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-central-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
}

func StreamVideo(context *gin.Context) {
	rangeHeader := context.Request.Header["Range"][0]
	ranges := strings.Split(strings.Replace(rangeHeader, "bytes=", "", 1), "-")
	start, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		context.Status(http.StatusNotAcceptable)
		context.Abort()
		return
	}
	// var end int64
	// if len(ranges) == 2 && len(ranges[1]) > 0 {
	// 	end, err = strconv.ParseInt(ranges[1], 10, 64)
	// 	if err != nil {
	// 		context.Status(http.StatusNotAcceptable)
	// 		context.Abort()
	// 		return
	// 	}
	// } else {
	// 	end = start + CHUNK_SIZE
	// }

	// binary read
	name := context.Param("name")
	file, err := os.Open("static/" + name + ".mp4")

	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		context.Abort()
		return
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		context.Abort()
		return
	}

	var totalFileSize int64 = stats.Size()
	end := start + CHUNK_SIZE
	if end > totalFileSize-1 {
		end = totalFileSize - 1
	}

	contentLength := end - start + 1
	data := make([]byte, contentLength)
	file.Seek(start, 0)
	bytesRead, _ := file.Read(data)
	fmt.Println("Bytes read: " + strconv.Itoa(bytesRead))

	context.Writer.Header().Add("Content-Range", "bytes "+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10)+"/"+strconv.FormatInt(totalFileSize, 10))
	context.Writer.Header().Add("Accept-Ranges", "bytes")
	context.Writer.Header().Add("Content-Length", strconv.FormatInt(contentLength, 10))
	context.Data(206, "video/mp4", data)
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

// func UploadVideo(c *gin.Context) {
// 	_, claims := utils.GetTokenClaims(c)
// 	if claims.Role != "RegisteredUser" {
// 		c.JSON(401, gin.H{"error": "unauthorized role"})
// 		c.Abort()
// 		return
// 	}

// 	// single file
// 	file, _ := c.FormFile("file")
// 	fmt.Println(file.Filename)

// 	if filepath.Ext(file.Filename) != ".mp4" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid extension"})
// 		c.Abort()
// 		return
// 	}

// 	// generate random number for filename
// 	rand.Seed(time.Now().UnixNano())
// 	rndNum := rand.Intn(math.MaxInt32-0) + 0
// 	filenameNoExt := strconv.Itoa(rndNum)
// 	file.Filename = filenameNoExt + ".mp4"

// 	// create record for video in db
// 	video := &models.Video{
// 		Title:       c.Query("title"),
// 		Description: c.Query("description"),
// 		OwnerEmail:  claims.Email,
// 		Filename:    filenameNoExt,
// 	}
// 	database.Instance.Save(&video)

// 	// Upload the file to specific dst.
// 	// wd, err := os.Getwd()
// 	// if err != nil {
// 	// 	panic(err)
// 	// }
// 	// parent := filepath.Dir(wd)
// 	err := c.SaveUploadedFile(file, "static/"+file.Filename)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}

// 	// generate thumbnail and save it to /static
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}
// 	videoFullPath := wd + "/static/" + file.Filename
// 	thumbnailOutputFile := generateVideoThumbnail(videoFullPath)
// 	thumbnailDst, err := os.Create("static/" + filenameNoExt + ".jpg")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create thumbnail file"})
// 		c.Abort()
// 		return
// 	}
// 	defer thumbnailDst.Close()
// 	io.Copy(thumbnailDst, thumbnailOutputFile)
// 	thumbnailOutputFile.Close()

// 	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
// }

// func UploadVideo(c *gin.Context) {
// 	_, claims := utils.GetTokenClaims(c)
// 	if claims.Role != "RegisteredUser" {
// 		c.JSON(401, gin.H{"error": "unauthorized role"})
// 		c.Abort()
// 		return
// 	}

// 	// single file
// 	file, _ := c.FormFile("file")
// 	fmt.Println(file.Filename)

// 	if filepath.Ext(file.Filename) != ".mp4" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid extension"})
// 		c.Abort()
// 		return
// 	}

// 	// Generate random filename
// 	rand.Seed(time.Now().UnixNano())
// 	rndNum := rand.Intn(math.MaxInt32-0) + 0
// 	filenameNoExt := strconv.Itoa(rndNum)
// 	file.Filename = filenameNoExt + ".mp4"

// 	// Upload file to S3
// 	src, err := file.Open()

// 	if err != nil {
// 		log.Println("Failed to open file:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
// 		return
// 	}
// 	defer src.Close()

// 	err = uploadToS3(src, file.Filename)
// 	if err != nil {
// 		log.Println("Failed to upload to S3:", err)
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to S3"})
// 		return
// 	}

// 	// generate thumbnail and save it to /static
// 	wd, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}
// 	videoFullPath := wd + "/static/" + file.Filename
// 	thumbnailOutputFile := generateVideoThumbnail(videoFullPath)
// 	thumbnailDst, err := os.Create("static/" + filenameNoExt + ".jpg")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create thumbnail file"})
// 		c.Abort()
// 		return
// 	}
// 	defer thumbnailDst.Close()
// 	io.Copy(thumbnailDst, thumbnailOutputFile)
// 	thumbnailOutputFile.Close()

// 	// Save video metadata to DB
// 	video := &models.Video{
// 		Title:       c.Query("title"),
// 		Description: c.Query("description"),
// 		OwnerEmail:  claims.Email,
// 		Filename:    filenameNoExt,
// 	}
// 	database.Instance.Save(&video)

// 	c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
// }

func UploadVideo(c *gin.Context) {
	_, claims := utils.GetTokenClaims(c)
	if claims.Role != "RegisteredUser" {
		c.JSON(401, gin.H{"error": "unauthorized role"})
		c.Abort()
		return
	}

	// single file
	file, _ := c.FormFile("file")
	fmt.Println(file.Filename)

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

	// Upload video file to S3
	src, err := file.Open()
	if err != nil {
		log.Println("Failed to open file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	err = uploadToS3(src, videoFilename, "video/mp4")
	if err != nil {
		log.Println("Failed to upload to S3:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to S3"})
		return
	}

	// Generate thumbnail and upload to S3
	videoFullPath := fmt.Sprintf("https://vide-oh-videos.s3.eu-central-1.amazonaws.com/%s", videoFilename)
	err = generateVideoThumbnail(videoFullPath, filenameNoExt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate thumbnail"})
		c.Abort()
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
		ACL:         types.ObjectCannedACLPublicRead,
	})
	return err
}

// func uploadToS3(file multipart.File, filename string) error {
// 	// Read file content
// 	buf := new(bytes.Buffer)
// 	buf.ReadFrom(file)

// 	// Upload to S3
// 	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
// 		Bucket:      aws.String(bucketName),
// 		Key:         aws.String(filename),
// 		Body:        bytes.NewReader(buf.Bytes()),
// 		ContentType: aws.String("video/mp4"),
// 		ACL:         types.ObjectCannedACLPrivate,
// 	})
// 	return err
// }

// func generateVideoThumbnail(url string) *os.File {
// 	tempDir, err := ioutil.TempDir("", "thumbnail*")
// 	if err != nil {
// 		panic(err)
// 	}

// 	outputFilePath := tempDir + "/thumbnail.png"

// 	cmd := `ffmpeg -i "%s" -an -q 0 -vf scale="'if(gt(iw,ih),-1,200):if(gt(iw,ih),200,-1)', crop=200:200:exact=1" -vframes 1 "%s"`
// 	// ffmpeg cmd ref : https://gist.github.com/TimothyRHuertas/b22e1a252447ab97aa0f8de7c65f96b8

// 	cmdSubstituted := fmt.Sprintf(cmd, url, outputFilePath)

// 	// shellName := "ash" // for docker (using alpine image)
// 	// if os.Getenv("ENV") != "" && os.Getenv("ENV") == "LOCAL" {
// 	// 	shellName = "bash"
// 	// }
// 	shellName := "bash"

// 	ffCmd := exec.Command(shellName, "-c", cmdSubstituted)

// 	// getting real error msg : https://stackoverflow.com/questions/18159704/how-to-debug-exit-status-1-error-when-running-exec-command-in-golang
// 	output, err := ffCmd.CombinedOutput()
// 	if err != nil {
// 		log.Println(fmt.Sprint(err) + ": " + string(output))
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// 	log.Println(string(output))

// 	outputFile, _ := os.Open(outputFilePath)
// 	return outputFile
// }

// func generateVideoThumbnail(url string) *os.File {
// 	// Use /tmp for temporary file storage in Lambda
// 	outputFilePath := "/tmp/thumbnail.png"

// 	cmd := `ffmpeg -i "%s" -an -q 0 -vf scale="'if(gt(iw,ih),-1,200):if(gt(iw,ih),200,-1)', crop=200:200:exact=1" -vframes 1 "%s"`
// 	cmdSubstituted := fmt.Sprintf(cmd, url, outputFilePath)

// 	// Use bash to execute the command
// 	shellName := "bash"
// 	ffCmd := exec.Command(shellName, "-c", cmdSubstituted)

// 	output, err := ffCmd.CombinedOutput()
// 	if err != nil {
// 		log.Println(fmt.Sprint(err) + ": " + string(output))
// 		panic(err)
// 	}
// 	log.Println(string(output))

// 	// Open the thumbnail file
// 	outputFile, err := os.Open(outputFilePath)
// 	if err != nil {
// 		log.Println("Failed to open thumbnail file:", err)
// 		panic(err)
// 	}
// 	return outputFile
// }

func generateVideoThumbnail(videoPath string, filenameNoExt string) error {
	// Use /tmp for temporary file storage in Lambda
	outputFilePath := fmt.Sprintf("/tmp/%s.png", filenameNoExt)

	cmd := fmt.Sprintf(`./ffmpeg -i "%s" -an -q 0 -vf scale='if(gt(iw,ih),-1,200):if(gt(iw,ih),200,-1)',crop=200:200:exact=1 -vframes 1 "%s"`, videoPath, outputFilePath)

	shellName := "bash"
	ffCmd := exec.Command(shellName, "-c", cmd)

	output, err := ffCmd.CombinedOutput()
	if err != nil {
		log.Println(fmt.Sprint(err) + ": " + string(output))
		return err
	}
	log.Println(string(output))

	// Upload thumbnail to S3
	file, err := os.Open(outputFilePath)
	if err != nil {
		log.Println("Failed to open thumbnail file:", err)
		return err
	}
	defer file.Close()

	return uploadToS3(file, fmt.Sprintf("%s.png", filenameNoExt), "image/png")
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
