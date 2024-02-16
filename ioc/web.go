package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
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
		}),

		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),

		(&middleware.LoginJWTMiddlewareBuilder{}).CheckLogin(),
	}
}
