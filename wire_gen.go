// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"webook/internal/events/article"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/web"
	"webook/internal/web/jwt"
	"webook/ioc"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitGinMiddlewares(cmdable, handler, logger)
	db := ioc.InitDB()
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewCachedUserRepository(userDAO, userCache)
	userService := service.NewUserService(userRepository)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeService(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, handler)
	database := ioc.InitMongoDB()
	node := ioc.InitSnowFlake()
	articleDAO := dao.NewMongoDBArticleDAO(database, node)
	articleCache := cache.NewArticleRedisCache(cmdable)
	articleRepository := repository.NewCachedArticleRepository(articleDAO, userRepository, articleCache)
	client := ioc.InitSaramaClient()
	syncProducer := ioc.InitSyncProducer(client)
	producer := article.NewSaramaSyncProducer(syncProducer)
	articleService := service.NewArticleService(articleRepository, producer)
	interactiveDAO := dao.NewGORMInteractiveDAO(db)
	interactiveCache := cache.NewInteractiveRedisCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepository(interactiveDAO, logger, interactiveCache)
	interactiveService := service.NewInteractiveService(interactiveRepository, producer, logger)
	articleHandler := web.NewArticleHandler(logger, articleService, interactiveService)
	wechatService := ioc.InitWechatService()
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	engine := ioc.InitWebServer(v, userHandler, articleHandler, oAuth2WechatHandler)
	interactiveReadEventConsumer := article.NewInteractiveReadEventConsumer(interactiveRepository, client, logger)
	interactiveLikeEventConsumer := article.NewInteractiveLikeEventConsumer(interactiveRepository, client, logger)
	interactiveUnLikeEventConsumer := article.NewInteractiveUnLikeEventConsumer(interactiveRepository, client, logger)
	v2 := ioc.InitConsumers(interactiveReadEventConsumer, interactiveLikeEventConsumer, interactiveUnLikeEventConsumer)
	monitorMessage := ioc.InitKafkaPrometheus(client)
	app := &App{
		server:       engine,
		consumers:    v2,
		kafkaMonitor: monitorMessage,
	}
	return app
}
