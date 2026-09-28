package service

import "abingblog-backend/internal/repository"

type SiteMetricService struct {
	repo      *repository.SiteMetricRepo
	visitRepo *repository.VisitLogRepo
}

func NewSiteMetricService(repo *repository.SiteMetricRepo, visitRepo *repository.VisitLogRepo) *SiteMetricService {
	return &SiteMetricService{repo: repo, visitRepo: visitRepo}
}

func (s *SiteMetricService) RecordStart() (uint, error) {
	return s.repo.IncrementStartCount()
}

// Total 只读总量 + 独立访客数：前端进屏幕展示用，不产生副作用
func (s *SiteMetricService) Total() (uint, int64, error) {
	count, err := s.repo.GetStartCount()
	if err != nil {
		return 0, 0, err
	}
	distinct, err := s.visitRepo.DistinctUsers()
	if err != nil {
		return 0, 0, err
	}
	return count, distinct, nil
}
