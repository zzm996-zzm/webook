package service

import (
	"webook/internal/domain"
	"webook/internal/repository"
)

type LikeService interface {
	GetTopNArticles(n int64) ([]domain.Article, error)
}

type likeService struct {
	intrRepo repository.InteractiveRepository
}

func NewLikeService(intrRepo repository.InteractiveRepository) LikeService {
	return &likeService{
		intrRepo: intrRepo,
	}
}

func (l *likeService) GetTopNArticles(n int64) ([]domain.Article, error) {
	// 先从本地缓存中获取数据
	return nil, nil
}
