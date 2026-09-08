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
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
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
