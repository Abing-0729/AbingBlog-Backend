package router

import (
	"abingblog-backend/internal/handler"
	"abingblog-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 注册所有路由。路由分组即权限边界：
//   - /api/v1 公共接口：游客可用，只暴露已发布内容
//   - /api/v1/admin 后台接口：挂 JWT 中间件，未登录一律 401
//
// CORS 中间件挂在最外层（gin.Default 之后、路由分组之前），保证预检 OPTIONS
// 和所有响应都带上跨域头；allowOrigins 是前端 dev server 白名单。
func Setup(h *handler.Handler, secret string, allowOrigins []string) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.Cors(allowOrigins))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/articles", h.Article.List)
		v1.GET("/articles/:id", h.Article.Get)
		v1.GET("/categories", h.Category.List)
		v1.GET("/tags", h.Tag.List)
		v1.GET("/healthz", h.Health.Check)
		v1.POST("/login", h.User.Login)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthRequired(secret))
		{
			admin.GET("/articles", h.Article.AdminList)
			admin.POST("/articles", h.Article.Create)
			admin.PUT("/articles/:id", h.Article.Update)
			admin.DELETE("/articles/:id", h.Article.Delete)
			admin.PUT("/articles/:id/status", h.Article.UpdateStatus)

			admin.POST("/categories", h.Category.Create)
			admin.PUT("/categories/:id", h.Category.Update)
			admin.DELETE("/categories/:id", h.Category.Delete)

			admin.POST("/tags", h.Tag.Create)
			admin.PUT("/tags/:id", h.Tag.Update)
			admin.DELETE("/tags/:id", h.Tag.Delete)

		}
	}
	return r
}
