package router

import (
	"abingblog-backend/internal/handler"
	"abingblog-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 注册所有路由。路由分组即权限边界：
//   - /api/v1 公共接口：游客可用，只暴露已发布内容
//   - /api/v1/admin 后台接口：下一迭代挂 JWT 中间件后才是真正私有
//   - TODO(JWT)：在 admin 分组挂鉴权中间件，未登录一律 401
//   - TODO(CORS)：前后端联调时给前端 dev server 加跨域白名单
func Setup(h *handler.Handler, secret string) *gin.Engine {
	r := gin.Default()

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
