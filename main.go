package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
	"webook/config"
	"webook/internal/repository"
	"webook/internal/repository/cache"
	"webook/internal/repository/dao"
	"webook/internal/service"
	"webook/internal/service/sms"
	"webook/internal/service/sms/localsms"
	"webook/internal/web"
	"webook/internal/web/middleware"
)

func initDB() *gorm.DB {
	// 初始化数据库连接
	db, err := gorm.Open(mysql.Open(config.Config.DB.DNS))
	if err != nil {
		panic(err)
	}

	// 初始化表结构
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}

	return db
}

func initWebServer() *gin.Engine {
	server := gin.Default()

	server.Use(cors.New(cors.Config{
		// 是否允许带上用户认证信息 比如cookie
		AllowCredentials: true,
		// 业务业务请求中可以带上的头
		AllowHeaders: []string{"Content-Type", "Authorization"},
		//允许前端访问自定义返回的token
		ExposeHeaders: []string{"x-jwt-token"},
		//哪些来源是允许的
		AllowOriginFunc: func(origin string) bool {
			if strings.Contains(origin, "localhost") {
				return false
			}
			return strings.Contains(origin, "zzm.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	//reids 限流
	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: config.Config.Redis.Addr,
	//})

	//server.Use(ratelimit.NewBuilder(redisClient, time.Second, 100).Build())

	UseJWT(server)

	return server

}

func initCodeSvc(redisClient redis.Cmdable) *service.CodeService {
	cc := cache.NewCodeCache(redisClient)
	crepo := repository.NewCodeRepository(cc)
	return service.NewCodeService(crepo, initMemorySms())
}

func initMemorySms() sms.Service {
	return localsms.NewService()
}

func UseJWT(server *gin.Engine) {
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

func UseSession(server *gin.Engine) {
	login := &middleware.LoginMiddlewareBuilder{}
	store := cookie.NewStore([]byte("secret"))
	//redis 实现 session
	//store, err := redis.NewStore(16, "tcp", "localhost:6379", "", []byte("00k2XQmvKq0uYdAtDy0msE6u6wpu8Fw0"), []byte("00k2XQmvKq0uYdAtDy0msE6u6wpu8Fw1"))
	//if err != nil {
	//	panic(err)
	//}

	server.Use(sessions.Sessions("ssid", store), login.CheckLogin())
}

func initUserHandler(db *gorm.DB, redisClient redis.Cmdable, codeSvc *service.CodeService, server *gin.Engine) {
	//注册用户模块
	ud := dao.NewUserDAO(db)
	uc := cache.NewUserCache(redisClient)
	ur := repository.NewUserRepository(ud, uc)
	us := service.NewUserService(ur)

	uh := web.NewUserHandler(us, codeSvc)
	uh.RegisterRoutes(server)
}

func main() {

	// 初始化DB
	db := initDB()
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})
	server := initWebServer()
	initCodeSvc := initCodeSvc(redisClient)

	initUserHandler(db, redisClient, initCodeSvc, server)

	//server := gin.Default()
	server.GET("/hello", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "hello world")
	})

	//启动服务
	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}
