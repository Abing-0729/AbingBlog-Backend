package repository

import (
	"abingblog-backend/internal/model"
	"gorm.io/gorm"
)

type SiteMetricRepo struct{ db *gorm.DB }

func NewSiteMetricRepo(db *gorm.DB) *SiteMetricRepo { return &SiteMetricRepo{db: db} }

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
