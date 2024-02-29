package repository

import (
	"context"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
)

type InteractiveRepository interface {
	IncrReadCnt(ctx context.Context, biz string, bizId int64) error
	IncrLike(ctx context.Context, biz string, id int64, uid int64) error
	DecrLike(ctx context.Context, biz string, id int64, uid int64) error
	AddCollectionItem(ctx context.Context, biz string, id int64, cid int64, uid int64) error
}

type CachedInteractiveRepository struct {
	dao   dao.InteractiveDAO
	cache cache.InteractiveCache
}

func (c *CachedInteractiveRepository) IncrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.InsertLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.IncrLikeCntIfPresent(ctx, biz, id)
}

func (c *CachedInteractiveRepository) DecrLike(ctx context.Context, biz string, id int64, uid int64) error {
	err := c.dao.DeleteLikeInfo(ctx, biz, id, uid)
	if err != nil {
		return err
	}
	return c.cache.DecrLikeCntIfPresent(ctx, biz, id)
}

func (repo *CachedInteractiveRepository) IncrReadCnt(ctx context.Context, biz string, bizId int64) error {
	// 先入库
	err := repo.dao.IncrReadCnt(ctx, biz, bizId)
	if err != nil {
		return err
	}

	return repo.cache.IncrReadCntIfPresent(ctx, biz, bizId)
}

func (c *CachedInteractiveRepository) AddCollectionItem(ctx context.Context,
	biz string, id int64, cid int64, uid int64) error {
	err := c.dao.InsertCollectionBiz(ctx, dao.UserCollectionBiz{
		Biz:   biz,
		BizId: id,
		Cid:   cid,
		Uid:   uid,
	})
	if err != nil {
		return err
	}
	return c.cache.IncrCollectCntIfPresent(ctx, biz, id)
}

func NewCachedInteractiveRepository(dao dao.InteractiveDAO, cache cache.InteractiveCache) InteractiveRepository {
	return &CachedInteractiveRepository{
		dao:   dao,
		cache: cache,
	}
}
