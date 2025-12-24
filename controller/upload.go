// controller/upload.go
package controller

import (
	"io"

	"github.com/X-Colder/companion-backend/service"
	"github.com/X-Colder/companion-backend/utils"

	"github.com/gin-gonic/gin"
)

// UploadController 文件上传控制器（仅处理Gin请求/响应，适配服务层）
type UploadController struct{}

// UploadAvatar 上传用户头像接口
func (u *UploadController) UploadAvatar(c *gin.Context) {
	// 1. 控制器层：用Gin接收前端上传的文件（表单字段名：avatar）
	file, err := c.FormFile("avatar")
	if err != nil {
		utils.Fail(c, "获取上传文件失败："+err.Error())
		return
	}

	// 2. 打开上传文件，获取读取流（返回*os.File，实现了io.ReadCloser接口）
	fileReader, err := file.Open()
	if err != nil {
		utils.Fail(c, "打开上传文件失败："+err.Error())
		return
	}
	defer fileReader.Close() // 延迟关闭文件读取流，此时fileReader是io.ReadCloser类型，可调用Close()

	// 3. 提取文件核心信息，调用服务层方法
	imgUrl, err := (&service.UploadService{}).UploadAvatar(file.Filename, file.Size, fileReader)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	// 4. 返回响应给前端
	utils.Success(c, gin.H{
		"img_url": imgUrl,
	})
}

// UploadEvalImgs 批量上传评价图片接口
func (u *UploadController) UploadEvalImgs(c *gin.Context) {
	// 1. 控制器层：用Gin接收批量上传的文件（表单字段名：eval_imgs）
	form, err := c.MultipartForm()
	if err != nil {
		utils.Fail(c, "获取批量上传文件失败："+err.Error())
		return
	}
	files := form.File["eval_imgs"]
	if len(files) == 0 {
		utils.Fail(c, "请选择要上传的图片")
		return
	}

	// 2. 构造服务层需要的文件信息切片（关键：FileReader改为io.ReadCloser类型）
	var fileInfos []struct {
		FileName   string
		FileSize   int64
		FileReader io.ReadCloser // 改为io.ReadCloser，支持Close()方法
	}
	for _, file := range files {
		// 打开每个文件，获取读取流（*os.File → 实现io.ReadCloser）
		reader, err := file.Open()
		if err != nil {
			// 关闭已打开的文件流，避免资源泄露
			for _, info := range fileInfos {
				info.FileReader.Close()
			}
			utils.Fail(c, "打开批量上传文件失败："+err.Error())
			return
		}
		fileInfos = append(fileInfos, struct {
			FileName   string
			FileSize   int64
			FileReader io.ReadCloser
		}{
			FileName:   file.Filename,
			FileSize:   file.Size,
			FileReader: reader,
		})
	}

	// 延迟关闭所有文件读取流（此时FileReader是io.ReadCloser类型，可正常调用Close()）
	defer func() {
		for _, info := range fileInfos {
			info.FileReader.Close()
		}
	}()

	// 3. 调用服务层批量上传方法（服务层接收io.Reader，io.ReadCloser可隐式转换为io.Reader）
	uploadFileInfos := make([]struct {
		FileName   string
		FileSize   int64
		FileReader io.Reader
	}, len(fileInfos))
	for i, info := range fileInfos {
		uploadFileInfos[i] = struct {
			FileName   string
			FileSize   int64
			FileReader io.Reader
		}{
			FileName:   info.FileName,
			FileSize:   info.FileSize,
			FileReader: info.FileReader,
		}
	}
	imgUrls, err := (&service.UploadService{}).UploadEvalImgs(uploadFileInfos)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	// 4. 返回响应给前端
	utils.Success(c, gin.H{
		"img_urls": imgUrls,
	})
}

// DeleteFile 删除单个文件接口
func (u *UploadController) DeleteFile(c *gin.Context) {
	// 1. 接收前端传递的文件访问路径
	var req struct {
		ImgUrl string `json:"img_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数格式错误："+err.Error())
		return
	}

	// 2. 调用服务层删除方法
	err := (&service.UploadService{}).DeleteFile(req.ImgUrl)
	if err != nil {
		utils.Fail(c, err.Error())
		return
	}

	// 3. 返回响应给前端
	utils.Success(c, "文件删除成功")
}
