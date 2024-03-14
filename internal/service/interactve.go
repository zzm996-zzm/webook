package service

import (
	"context"
	"golang.org/x/sync/errgroup"
	"webook/internal/domain"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/pkg/logger"
)

type InteractiveService interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	Like(c context.Context, biz string, id int64, uid int64) error
	CancelLike(c context.Context, biz string, id int64, uid int64) error
	Collect(ctx context.Context, biz string, bizId, cid, uid int64) error
	Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error)
}

type interactiveService struct {
	repo     repository.InteractiveRepository
	producer article.Producer
	l        logger.Logger
}

func (i *interactiveService) Get(ctx context.Context, biz string, id int64, uid int64) (domain.Interactive, error) {
	intr, err := i.repo.Get(ctx, biz, id)
	if err != nil {
		return domain.Interactive{}, err
	}

	var eg errgroup.Group

	// 查询该用户是否点赞
	eg.Go(func() error {
		var er error
		intr.Liked, er = i.repo.Liked(ctx, biz, id, uid)
		return er
	})

	// 查询该用户是否收藏
	eg.Go(func() error {
		var er error
		intr.Collected, er = i.repo.Collected(ctx, biz, id, uid)
		return er
	})

	return intr, eg.Wait()
}

func (i *interactiveService) Like(c context.Context, biz string, id int64, uid int64) error {
	//err := i.repo.IncrLike(c, biz, id, uid)
	go func() {
		// 发送消息

		evt := article.LikeEvent{
			Aid: id,
			Uid: uid,
		}

		err := i.producer.ProducerLikeEvent(evt)

		if err != nil {
			i.l.Error("发送 LikeEvent 失败",
				logger.Int64("aid", id),
				logger.Int64("uid", uid),
				logger.Error(err))
		}

	}()

	return nil
}

func (i *interactiveService) CancelLike(c context.Context, biz string, id int64, uid int64) error {
	// 同步发送
	//err := i.repo.DecrLike(c, biz, id, uid)
	// 异步消息队列发送
	go func() {
		// 发送消息

		evt := article.UnLikeEvent{
			Aid: id,
			Uid: uid,
		}

		err := i.producer.ProducerUnLikeEvent(evt)

		if err != nil {
			i.l.Error("发送 UnLikeEvent 失败",
				logger.Int64("aid", id),
				logger.Int64("uid", uid),
				logger.Error(err))
		}

	}()

	return nil
}

func (i *interactiveService) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	return i.repo.IncrReadCnt(ctx, biz, bizId)
}

func (i *interactiveService) Collect(ctx context.Context, biz string, bizId, cid, uid int64) error {
	return i.repo.AddCollectionItem(ctx, biz, bizId, cid, uid)
}

func NewInteractiveService(repo repository.InteractiveRepository, producer article.Producer, l logger.Logger) InteractiveService {
	return &interactiveService{repo: repo, producer: producer, l: l}
}
