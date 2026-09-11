package main

import (
	"fmt"
	"log"
	"os"

	"abingblog-backend/internal/config"
	"abingblog-backend/internal/database"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

func main() {
	cfg := config.Load()
	db := database.Init(cfg)
	repo := repository.NewUserRepo(db)

	username := envOrDefault("ABINGBLOG_ADMIN_USERNAME", "admin")
	password := envOrDefault("ABINGBLOG_ADMIN_PASSWORD", "admin123")
	if username == "" || password == "" {
		log.Fatal("管理员用户名和密码不能为空")
	}

	existing, err := repo.GetByUsername(username)
	if err != nil {
		log.Fatalf("查询管理员失败: %v", err)
	}
	if existing != nil {
		fmt.Printf("管理员 %q 已存在，跳过创建\n", username)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		log.Fatalf("生成密码哈希失败: %v", err)
	}
	user := &model.User{Username: username, Password: string(hash)}
	if err := repo.Create(user); err != nil {
		log.Fatalf("创建管理员失败: %v", err)
	}
	fmt.Printf("管理员 %q 创建成功（id=%d）\n", username, user.ID)
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
