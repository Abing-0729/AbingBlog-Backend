package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config 全局配置结构，对应 configs/config.yaml
type Config struct {
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Admin  AdminConfig  `mapstructure:"admin"`
	CORS   CORSConfig   `mapstructure:"cors"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// RedisConfig 缓存连接配置。生产环境用环境变量覆盖 host/password
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"` // 无密码则留空
	DB       int    `mapstructure:"db"`       // 逻辑库编号，默认 0
}

// Addr 拼接 go-redis 需要的 "host:port"
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// CORSConfig 跨域白名单。允许的前端来源列表，生产环境用环境变量覆盖
type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

// JWTConfig 签发/校验 token 的配置
type JWTConfig struct {
	Secret      string `mapstructure:"secret"`       // 签名密钥，生产用环境变量覆盖，绝不用默认值
	ExpireHours int    `mapstructure:"expire_hours"` // token 有效期（小时）
}

// AdminConfig 初始管理员账号。启动 seed 用，账号已存在则跳过；生产用环境变量注入
type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// MySQLConfig 数据库连接配置
type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
}

// DSN 拼接 go-sql-driver/mysql 的连接串（parseTime 让 time.Time 直接映射 DATETIME）
func (m MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.DBName, m.Charset)
}

// Load 读取 configs/config.yaml；viper.AutomaticEnv 支持环境变量覆盖，为 Docker 部署做准备
func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}
	return &cfg
}
