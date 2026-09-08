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
	cfg := config.Load()

	db := database.Init(cfg) // 连接 MySQL（自动建库建表）

	// 依赖注入，调用链：HTTP 请求 → handler → service → repository → GORM → MySQL
	articleRepo := repository.NewArticleRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	tagRepo := repository.NewTagRepo(db)

	h := &handler.Handler{
		Article:  handler.NewArticleHandler(service.NewArticleService(articleRepo, categoryRepo, tagRepo)),
		Category: handler.NewCategoryHandler(service.NewCategoryService(categoryRepo)),
		Tag:      handler.NewTagHandler(service.NewTagService(tagRepo)),
		Health:   handler.NewHealthHandler(db),
	}

	r := router.Setup(h)
	log.Printf("server running at http://localhost:%d", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatal(err)
	}
}
