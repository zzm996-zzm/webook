package service

import (
	"context"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type ArticleService interface {
	Save(ctx context.Context, art domain.Article) (int64, error)
	Publish(ctx context.Context, art domain.Article) (int64, error)
	Withdraw(ctx context.Context, uid int64, id int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id, uid int64) (domain.Article, error)
}

type articleService struct {
	repo     repository.ArticleRepository
	producer article.Producer
	l        logger.Logger
}

func (a *articleService) GetPubById(ctx context.Context, id, uid int64) (domain.Article, error) {
	res, err := a.repo.GetPubById(ctx, id)
	go func() {
		// 发送消息
		if err == nil {
			evt := article.ReadEvent{
				Aid: id,
				Uid: uid,
			}

			err := a.producer.ProducerReadEvent(evt)

			if err != nil {
				a.l.Error("发送 ReadEvent 失败",
					logger.Int64("aid", id),
					logger.Int64("uid", uid),
					logger.Error(err))
			}
		}
	}()

	return res, err
}

func (a *articleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, id)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, art)
}

func NewArticleService(repo repository.ArticleRepository, producer article.Producer) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
	}
}

func (a *articleService) Withdraw(ctx context.Context, uid int64, id int64) error {
	return a.repo.SyncStatus(ctx, uid, id, domain.ArticleStatusPrivate)
}

func (a *articleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	art.Status = domain.ArticleStatusUnpublished
	if art.Id > 0 {
		err := a.repo.Update(ctx, art)
		return art.Id, err
	}
	return a.repo.Create(ctx, art)
}
