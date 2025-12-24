package conf

import (
	"log"

	"github.com/spf13/viper"
)

// AppConfig 全局配置结构体
var AppConfig = &Config{}

// Config 配置结构体（对应app.yaml）
type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	Mysql struct {
		Dsn             string `mapstructure:"dsn"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"mysql"`
	Jwt struct {
		Secret      string `mapstructure:"secret"`
		ExpireHours int    `mapstructure:"expire_hours"`
	} `mapstructure:"jwt"`
	Upload struct {
		BasePath string   `mapstructure:"base_path"`
		MaxSize  int64    `mapstructure:"max_size"`
		AllowExt []string `mapstructure:"allow_ext"`
	} `mapstructure:"upload"`
}

// LoadConfig 加载配置文件
func LoadConfig() {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./conf")

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败：%s", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("解析配置失败：%s", err)
	}

	log.Println("配置文件加载成功")
}
