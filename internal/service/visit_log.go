package service

import (
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// VisitLogService 访客记录业务层：目前主要是查询与分页兜底
type VisitLogService struct {
	repo *repository.VisitLogRepo
}

func NewVisitLogService(repo *repository.VisitLogRepo) *VisitLogService {
	return &VisitLogService{repo: repo}
}

func (s *VisitLogService) List(page, pageSize int, keyword string) ([]model.VisitLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.repo.List(repository.VisitQuery{Page: page, PageSize: pageSize, Keyword: keyword})
}

// Summary 汇总数字 + 最近访客 + 热门接口
func (s *VisitLogService) Summary() (repository.VisitStats, []repository.VisitorSummary, []repository.TopPath, error) {
	stats, err := s.repo.Stats()
	if err != nil {
		return stats, nil, nil, err
	}
	visitors, err := s.repo.SummarizeByVisitor(20)
	if err != nil {
		return stats, nil, nil, err
	}
	paths, err := s.repo.TopPaths(5)
	return stats, visitors, paths, err
}
