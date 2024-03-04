package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
	"webook/internal/web"
	ijwt "webook/internal/web/jwt"
	"webook/internal/web/middleware"
	"webook/pkg/ginx/middleware/prometheus"
	"webook/pkg/ginx/middleware/ratelimit"
	"webook/pkg/limiter"
	"webook/pkg/logger"
)

func InitWebServer(mdls []gin.HandlerFunc,
	userHdl *web.UserHandler,
	artHdl *web.ArticleHandler,
	wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mdls...)
	userHdl.RegisterRoutes(server)
	wechatHdl.RegisterRoutes(server)
	artHdl.RegisterRoutes(server)
	return server
}

func InitGinMiddlewares(redisClient redis.Cmdable, hdl ijwt.Handler, l logger.Logger) []gin.HandlerFunc {
	pb := &prometheus.Builder{
		Namespace: "zzm",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 GIN 的HTTP接口数据",
	}
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			// 是否允许带上用户认证信息 比如cookie
			AllowCredentials: true,
			// 业务业务请求中可以带上的头
			AllowHeaders: []string{"Content-Type", "Authorization"},
			//允许前端访问自定义返回的token
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
			//哪些来源是允许的
			AllowOriginFunc: func(origin string) bool {
				if strings.Contains(origin, "localhost") {
					return false
				}
				return strings.Contains(origin, "zzm.com")
			},
			MaxAge: 12 * time.Hour,
		}),

		pb.BuilderResponseTime(),
		pb.BuilderActiveRequest(),

		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),

		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("middleware", logger.Field{Key: "req", Val: al})
		}).AllowReqBody().AllowRespBody().Build(),

		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("middleware", logger.Field{Key: "error", Val: al})
		}).AllowReqBody().AllowRespBody().CatchError(),

		middleware.NewLoginJWTMiddlewareBuilder(hdl).CheckLogin(),
	}
}
