package handler

// Handler 聚合所有接口处理器，由 main 统一组装依赖后传给 router
type Handler struct {
	Article    *ArticleHandler
	Category   *CategoryHandler
	Tag        *TagHandler
	Health     *HealthHandler
	Auth       *AuthHandler
	SiteMetric *SiteMetricHandler
}
