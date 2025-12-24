package service

import (
	"errors"

	"github.com/X-Colder/companion-backend/conf"
	"github.com/X-Colder/companion-backend/model" // 必须导入 model 包，才能访问 model.DB
	"github.com/X-Colder/companion-backend/utils"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct{}

// Login 登录逻辑
func (u *UserService) Login(phone, password string) (string, *model.User, error) {
	// 查询用户：修正为 model.DB（全局数据库实例）
	var user model.User
	// 第21行修正：gorm.DB → model.DB
	if err := model.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return "", nil, errors.New("手机号不存在")
		}
		return "", nil, errors.New("查询用户失败")
	}

	// 校验密码（bcrypt比对）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.UserType, conf.AppConfig.Jwt.Secret, conf.AppConfig.Jwt.ExpireHours)
	if err != nil {
		return "", nil, errors.New("生成token失败")
	}

	return token, &user, nil
}

// Register 注册逻辑
func (u *UserService) Register(phone string, userType int, password string) error {
	// 检查手机号是否已存在：修正为 model.DB
	var existUser model.User
	if err := model.DB.Where("phone = ?", phone).First(&existUser).Error; err == nil {
		return errors.New("手机号已注册")
	} else if !gorm.IsRecordNotFoundError(err) {
		return errors.New("查询用户失败")
	}

	// 密码加密（bcrypt）
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 创建用户：修正为 model.DB
	user := model.User{
		Phone:    phone,
		Password: string(hashPwd),
		UserType: userType,
	}
	if err := model.DB.Create(&user).Error; err != nil {
		return errors.New("注册失败")
	}

	return nil
}

// UpdateProfile 修改个人资料
func (u *UserService) UpdateProfile(userId uint64, nickname, avatar string) error {
	// 构造更新参数
	updateData := map[string]interface{}{
		"nickname": nickname,
	}
	if avatar != "" {
		updateData["avatar"] = avatar
	}

	// 更新用户：修正为 model.DB
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).Updates(updateData).Error; err != nil {
		return errors.New("修改资料失败")
	}

	return nil
}

// GetUserInfo 获取用户信息
func (u *UserService) GetUserInfo(userId uint64) (*model.User, error) {
	var user model.User
	// 查询用户：修正为 model.DB
	if err := model.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("查询用户失败")
	}

	return &user, nil
}
