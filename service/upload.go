// service/upload.go
package service

import (
	"errors"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/X-Colder/companion-backend/utils"
)

// UploadService 文件上传服务（完全解耦Gin，仅依赖Go内置包）
type UploadService struct{}

// 定义上传常量
const (
	// 允许的图片格式
	AllowImgExts = ".jpg,.jpeg,.png,.gif,.webp"
	// 单个文件最大大小（2MB）
	MaxFileSize = 2 * 1024 * 1024
	// 基础存储目录（相对于项目根目录）
	BaseUploadPath = "./static/upload"
)

// -------------------------- 通用上传方法（内部私有化） --------------------------
// uploadFile 通用文件保存逻辑（纯内置包实现）
// 参数：
//
//	fileName - 原始文件名（用于获取文件格式）
//	fileSize - 文件大小（用于校验）
//	fileReader - 文件读取流（用于读取文件内容保存到本地）
//	saveSubDir - 业务子目录（如avatar/eval）
func (u *UploadService) uploadFile(fileName string, fileSize int64, fileReader io.Reader, saveSubDir string) (string, error) {
	// 1. 校验文件大小
	if fileSize <= 0 || fileSize > MaxFileSize {
		return "", errors.New("文件大小无效或超过限制，最大支持2MB")
	}

	// 2. 校验文件格式
	fileExt := strings.ToLower(path.Ext(fileName))
	if !strings.Contains(AllowImgExts, fileExt) {
		return "", errors.New("不支持的文件格式，仅允许：" + AllowImgExts)
	}

	// 3. 构造存储路径
	dateDir := time.Now().Format("20060102")                                   // 按日期分目录，避免单目录文件过多
	fullSubDir := path.Join(saveSubDir, dateDir)                               // 业务子目录 + 日期目录
	saveDir := path.Join(BaseUploadPath, fullSubDir)                           // 本地完整存储目录
	uniqueFileName := utils.GetRandomString(16) + fileExt                      // 生成唯一文件名，避免冲突
	fullFilePath := path.Join(saveDir, uniqueFileName)                         // 本地完整文件路径
	accessPath := "/" + path.Join("static/upload", fullSubDir, uniqueFileName) // 前端访问路径

	// 4. 自动创建不存在的目录（权限0755：所有者可读可写可执行，其他用户可读可执行）
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", errors.New("创建存储目录失败：" + err.Error())
	}

	// 5. 创建本地文件（权限0644：所有者可读可写，其他用户可读）
	dstFile, err := os.Create(fullFilePath)
	if err != nil {
		return "", errors.New("创建本地文件失败：" + err.Error())
	}
	defer dstFile.Close() // 延迟关闭文件句柄，避免资源泄露

	// 6. 读取上传文件内容，写入本地文件
	_, err = io.Copy(dstFile, fileReader)
	if err != nil {
		return "", errors.New("保存文件内容失败：" + err.Error())
	}

	return accessPath, nil
}

// -------------------------- 业务专属上传方法（对外暴露） --------------------------
// UploadAvatar 上传用户头像
func (u *UploadService) UploadAvatar(fileName string, fileSize int64, fileReader io.Reader) (string, error) {
	return u.uploadFile(fileName, fileSize, fileReader, "avatar")
}

// UploadEvalImg 单张上传评价图片
func (u *UploadService) UploadEvalImg(fileName string, fileSize int64, fileReader io.Reader) (string, error) {
	return u.uploadFile(fileName, fileSize, fileReader, "eval")
}

// UploadEvalImgs 批量上传评价图片（返回所有图片的访问路径）
func (u *UploadService) UploadEvalImgs(fileInfos []struct {
	FileName   string
	FileSize   int64
	FileReader io.Reader
}) ([]string, error) {
	var imgUrls []string
	for _, fileInfo := range fileInfos {
		url, err := u.UploadEvalImg(fileInfo.FileName, fileInfo.FileSize, fileInfo.FileReader)
		if err != nil {
			return nil, err // 批量上传失败则终止，可根据需求改为跳过失败文件
		}
		imgUrls = append(imgUrls, url)
	}
	return imgUrls, nil
}

// -------------------------- 文件删除方法 --------------------------
// DeleteFile 根据前端访问路径删除本地文件
func (u *UploadService) DeleteFile(accessPath string) error {
	// 前端访问路径转本地文件路径（去掉开头的/，拼接基础目录）
	localPath := "." + accessPath

	// 校验文件是否存在
	_, err := os.Stat(localPath)
	if os.IsNotExist(err) {
		return errors.New("文件不存在，无需删除")
	}
	if err != nil {
		return errors.New("查询文件状态失败：" + err.Error())
	}

	// 删除本地文件
	if err := os.Remove(localPath); err != nil {
		return errors.New("删除文件失败：" + err.Error())
	}
	return nil
}

// BatchDeleteFiles 批量删除文件
func (u *UploadService) BatchDeleteFiles(accessPaths []string) error {
	for _, path := range accessPaths {
		if err := u.DeleteFile(path); err != nil {
			return err
		}
	}
	return nil
}
