package controller

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/X-Colder/companion-backend/conf"
	"github.com/X-Colder/companion-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// UploadController 文件上传控制器
type UploadController struct{}

// 初始化MinIO客户端
var minioClient *minio.Client

func init() {
	var err error
	minioClient, err = minio.New(conf.MinIOEndpoint, &minio.Options{
		Creds:  minio.NewStaticV4(conf.MinIOAccessKeyID, conf.MinIOSecretAccessKey, ""),
		Secure: conf.MinIOUseSSL,
	})
	if err != nil {
		panic("MinIO客户端初始化失败：" + err.Error())
	}
	// 检查桶是否存在，不存在则创建
	exists, err := minioClient.BucketExists(context.Background(), conf.MinIOBucketName)
	if err != nil {
		panic("检查MinIO桶失败：" + err.Error())
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), conf.MinIOBucketName, minio.MakeBucketOptions{})
		if err != nil {
			panic("创建MinIO桶失败：" + err.Error())
		}
	}
}

// UploadImg 上传图片
func (u *UploadController) UploadImg(c *gin.Context) {
	// 获取上传文件
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "获取图片失败：" + err.Error(),
			"data": nil,
		})
		return
	}
	defer file.Close()

	// 生成唯一文件名
	ext := filepath.Ext(fileHeader.Filename)
	fileName := "img/" + utils.GenerateUUID() + ext

	// 上传到MinIO
	_, err = minioClient.PutObject(context.Background(), conf.MinIOBucketName, fileName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "图片上传失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	// 获取文件访问地址（带有效期）
	reqParams := make(map[string]string)
	presignedURL, err := minioClient.PresignedGetObject(context.Background(), conf.MinIOBucketName, fileName, time.Hour*24*365, reqParams)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "获取图片地址失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "图片上传成功",
		"data": gin.H{
			"url": presignedURL.String(),
		},
	})
}

// UploadVideo 上传视频（逻辑与图片类似，仅修改ContentType和文件路径）
func (u *UploadController) UploadVideo(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "获取视频失败：" + err.Error(),
			"data": nil,
		})
		return
	}
	defer file.Close()

	ext := filepath.Ext(fileHeader.Filename)
	fileName := "video/" + utils.GenerateUUID() + ext

	_, err = minioClient.PutObject(context.Background(), conf.MinIOBucketName, fileName, file, fileHeader.Size, minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "视频上传失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	reqParams := make(map[string]string)
	presignedURL, err := minioClient.PresignedGetObject(context.Background(), conf.MinIOBucketName, fileName, time.Hour*24*365, reqParams)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 500,
			"msg":  "获取视频地址失败：" + err.Error(),
			"data": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "视频上传成功",
		"data": gin.H{
			"url": presignedURL.String(),
		},
	})
}
