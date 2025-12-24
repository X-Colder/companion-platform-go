package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/X-Colder/companion-backend/conf"
	"github.com/X-Colder/companion-backend/model"
	"github.com/X-Colder/companion-backend/router"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func main() {
	// 加载配置
	conf.LoadConfig()

	// 初始化数据库连接
	initDB()

	// 初始化路由
	r := router.InitRouter()

	// 启动服务
	port := ":" + conf.AppConfig.Server.Port
	srv := &http.Server{
		Addr:    port,
		Handler: r,
	}

	// 异步启动服务
	go func() {
		log.Printf("服务启动成功，监听端口：%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败：%s", err)
		}
	}()

	// 优雅关闭服务（监听信号）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("开始关闭服务...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("服务关闭失败：%s", err)
	}
	log.Println("服务已关闭")
}

// initDB 初始化数据库连接
func initDB() {
	// 连接MySQL
	db, err := gorm.Open("mysql", conf.AppConfig.Mysql.Dsn)
	if err != nil {
		log.Fatalf("数据库连接失败：%s", err)
	}

	// 数据库配置
	db.DB().SetMaxIdleConns(conf.AppConfig.Mysql.MaxIdleConns)
	db.DB().SetMaxOpenConns(conf.AppConfig.Mysql.MaxOpenConns)
	db.DB().SetConnMaxLifetime(time.Duration(conf.AppConfig.Mysql.ConnMaxLifetime) * time.Second)

	// 开启日志（debug模式）
	if conf.AppConfig.Server.Mode == "debug" {
		db.LogMode(true)
	}

	// 自动迁移表（不存在则创建，存在则不修改结构）
	db.AutoMigrate(
		&model.User{},
		&model.Demand{},
		&model.Order{},
		&model.Evaluation{},
		&model.BalanceRecord{},
	)

	// 全局保存DB实例
	model.DB = db
	log.Println("数据库初始化成功")
}
