// model/db.go
package model

import "github.com/jinzhu/gorm"

// DB 全局GORM数据库实例（由main.go初始化后赋值）
var DB *gorm.DB
