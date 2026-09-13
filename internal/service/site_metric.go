package service

import "abingblog-backend/internal/repository"

type SiteMetricService struct{ repo *repository.SiteMetricRepo }

func NewSiteMetricService(repo *repository.SiteMetricRepo) *SiteMetricService {
	return &SiteMetricService{repo: repo}
}

func (s *SiteMetricService) RecordStart() (uint, error) {
	return s.repo.IncrementStartCount()
}
