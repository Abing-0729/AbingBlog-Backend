package repository

import (
	"errors"

	"abingblog-backend/internal/model"
	"gorm.io/gorm"
)

type SiteMetricRepo struct{ db *gorm.DB }

func NewSiteMetricRepo(db *gorm.DB) *SiteMetricRepo { return &SiteMetricRepo{db: db} }

// GetStartCount 只读总量：绝不自增，也不建行。
// 计数行还没被创建过（从没人点过 START）时返回 0，而不是报错——
// 前端进屏幕就调这个，冷启动也要能拿到一个干净的 0。
func (r *SiteMetricRepo) GetStartCount() (uint, error) {
	var metric model.SiteMetric
	err := r.db.First(&metric, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return metric.StartCount, nil
}

func (r *SiteMetricRepo) IncrementStartCount() (uint, error) {
	if err := r.db.Where("id = ?", 1).FirstOrCreate(&model.SiteMetric{ID: 1}).Error; err != nil {
		return 0, err
	}
	if err := r.db.Model(&model.SiteMetric{}).Where("id = ?", 1).
		UpdateColumn("start_count", gorm.Expr("start_count + ?", 1)).Error; err != nil {
		return 0, err
	}
	var metric model.SiteMetric
	if err := r.db.First(&metric, 1).Error; err != nil {
		return 0, err
	}
	return metric.StartCount, nil
}
