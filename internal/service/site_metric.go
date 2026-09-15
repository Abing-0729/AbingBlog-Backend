package service

import "abingblog-backend/internal/repository"

type SiteMetricService struct{ repo *repository.SiteMetricRepo }

func NewSiteMetricService(repo *repository.SiteMetricRepo) *SiteMetricService {
	return &SiteMetricService{repo: repo}
}

func (s *SiteMetricService) RecordStart() (uint, error) {
	return s.repo.IncrementStartCount()
}

// Total 只读总量：前端进屏幕展示用，不产生副作用
func (s *SiteMetricService) Total() (uint, error) {
	return s.repo.GetStartCount()
}
