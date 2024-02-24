//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/ioc"
)

var thirdPartySet = wire.NewSet( // 第三方依赖
	InitRedis, InitDB,
	InitLogger)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdPartySet,
		// DAO 部分
		dao.NewUserDAO,
		dao.NewArticleGORMDAO,

		// cache 部分
		cache.NewCodeCache, cache.NewUserCache,
		cache.NewArticleRedisCache,

		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,
		repository.NewCachedArticleRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		InitWechatService,

		// handler 部分
		web.NewUserHandler,
		web.NewArticleHandler,
		web.NewOAuth2WechatHandler,
		ijwt.NewRedisJWTHandler,
		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}

func InitArticleHandler(dao dao.ArticleDAO) *web.ArticleHandler {
	wire.Build(
		thirdPartySet,
		service.NewArticleService,
		web.NewArticleHandler,
		cache.NewArticleRedisCache,
		repository.NewCachedArticleRepository)
	return &web.ArticleHandler{}
}
