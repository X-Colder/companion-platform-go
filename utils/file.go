package utils

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadFile 文件上传工具
// fileKey: 前端上传的文件字段名（如"file"）
// saveDir: 存储子目录（如"avatar"/"eval"）
// basePath: 基础存储路径（如"./static/upload/"）
// maxSize: 最大文件大小（MB）
// allowExt: 允许的文件后缀
func UploadFile(c *gin.Context, fileKey, saveDir, basePath string, maxSize int64, allowExt []string) (string, error) {
	// 获取上传文件
	file, fileHeader, err := c.Request.FormFile(fileKey)
	if err != nil {
		return "", errors.New("获取上传文件失败")
	}
	defer file.Close()

	// 校验文件大小
	fileSize := fileHeader.Size
	if fileSize > maxSize*1024*1024 {
		return "", fmt.Errorf("文件大小超过%dMB限制", maxSize)
	}

	// 校验文件后缀
	fileExt := strings.ToLower(path.Ext(fileHeader.Filename))
	allow := false
	for _, ext := range allowExt {
		if fileExt == ext {
			allow = true
			break
		}
	}
	if !allow {
		return "", errors.New("文件格式不允许，请上传图片格式文件")
	}

	// 构造文件路径
	// 按日期分目录（如：2025/12/24/）
	dateDir := time.Now().Format("2006/01/02")
	fullSaveDir := path.Join(basePath, saveDir, dateDir)
	// 创建目录（不存在则创建）
	if err := os.MkdirAll(fullSaveDir, 0755); err != nil {
		return "", errors.New("创建存储目录失败")
	}

	// 生成唯一文件名（UUID+后缀）
	fileName := uuid.New().String() + fileExt
	fullFilePath := path.Join(fullSaveDir, fileName)
	// 拼接访问URL（前端可直接访问）
	accessUrl := fmt.Sprintf("/static/upload/%s/%s/%s", saveDir, dateDir, fileName)

	// 创建文件
	dstFile, err := os.Create(fullFilePath)
	if err != nil {
		return "", errors.New("创建文件失败")
	}
	defer dstFile.Close()

	// 写入文件
	_, err = dstFile.ReadFrom(file)
	if err != nil {
		return "", errors.New("写入文件失败")
	}

	return accessUrl, nil
}
