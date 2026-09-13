package main

import (
	"fmt"
	"log"

	"abingblog-backend/internal/config"
	"abingblog-backend/internal/database"
	"abingblog-backend/internal/handler"
	"abingblog-backend/internal/repository"
	"abingblog-backend/internal/router"
	"abingblog-backend/internal/service"
)

func main() {
	// 启动配置
	cfg := config.Load()

	db := database.Init(cfg) // 连接 MySQL（自动建库建表）

	// 依赖注入，调用链：HTTP 请求 → handler → service → repository → GORM → MySQL
	articleRepo := repository.NewArticleRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	tagRepo := repository.NewTagRepo(db)
	userRepo := repository.NewUserRepo(db)

	userSvc := service.NewUserService(userRepo, cfg)
	// 启动时确保初始管理员存在（账号已存在则跳过）
	if err := userSvc.SeedAdmin(); err != nil {
		log.Fatalf("初始化管理员失败: %v", err)
	}

	// 拦截链启动
	h := &handler.Handler{
		Article:  handler.NewArticleHandler(service.NewArticleService(articleRepo, categoryRepo, tagRepo)),
		Category: handler.NewCategoryHandler(service.NewCategoryService(categoryRepo)),
		Tag:      handler.NewTagHandler(service.NewTagService(tagRepo)),
		User:     handler.NewUserHandler(userSvc),
		Health:   handler.NewHealthHandler(db),
	}

	// 启动
	r := router.Setup(h, cfg.JWT.Secret)
	log.Printf("server running at http://localhost:%d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatal(err)
	}
}
