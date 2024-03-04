//go:build wireinject

package main

import (
	"github.com/google/wire"
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

func InitWebServer() *App {
	wire.Build(
		// 第三方依赖
		ioc.InitSaramaClient,
		ioc.InitRedis, ioc.InitDB,
		ioc.InitLogger,
		ioc.InitMongoDB,
		ioc.InitSnowFlake,
		ioc.InitSyncProducer,
		ioc.InitConsumers,
		article.NewInteractiveReadEventConsumer,
		article.NewSaramaSyncProducer,

		// DAO 部分
		dao.NewUserDAO,
		//dao.NewArticleGORMDAO,
		dao.NewMongoDBArticleDAO,
		dao.NewGORMInteractiveDAO,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,
		cache.NewInteractiveRedisCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,
		repository.NewCachedInteractiveRepository,

		// Service 部分
		ioc.InitSMSService,
		ioc.InitWechatService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		ijwt.NewRedisJWTHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
