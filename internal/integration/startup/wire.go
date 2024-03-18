//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/async"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

var userSvcProvider = wire.NewSet(
	dao.NewUserDAO,
	cache.NewUserCache,
	repository.NewCachedUserRepository,
	service.NewUserService)

var articlSvcProvider = wire.NewSet(
	repository.NewCachedArticleRepository,
	cache.NewArticleRedisCache,
	dao.NewArticleGORMDAO,
	service.NewArticleService)

//func InitWebServer() *gin.Engine {
//	wire.Build(
//		thirdPartySet,
//		userSvcProvider,
//		articlSvcProvider,
//		// cache 部分
//		cache.NewCodeCache,
//
//		// repository 部分
//		repository.NewCodeRepository,
//
//		// Service 部分
//		ioc.InitSMSService,
//		service.NewCodeService,
//		InitWechatService,
//
//		// handler 部分
//		web.NewUserHandler,
//		web.NewArticleHandler,
//		web.NewOAuth2WechatHandler,
//		ijwt.NewRedisJWTHandler,
//		ioc.InitGinMiddlewares,
//		ioc.InitWebServer,
//	)
//	return gin.Default()
//}

func InitAsyncSmsService(svc sms.Service) *async.Service {
	wire.Build(thirdPartySet, repository.NewAsyncSmSRepository,
		dao.NewGORMAsyncSmsDAO,
		async.NewService,
		async.NewOptions,
	)
	return &async.Service{}
}

//func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
//	wire.Build(
//		thirdPartySet,
//		userSvcProvider,
//		repository.NewCachedArticleRepository,
//		cache.NewArticleRedisCache,
//		service.NewArticleService,
//		web.NewArticleHandler)
//	return &web.ArticleHandler{}
//}
