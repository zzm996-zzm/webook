package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	ijwt "webook/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			// 不需要登录校验
			return
		}

		tokenStr := m.ExtractToken(ctx)

		var uc = ijwt.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			return
		}

		if token == nil || !token.Valid {
			return
		}

		// user-agent不同
		if ctx.GetHeader("User-Agent") != uc.UserAgent {
			// 后期监控告警的时候埋点
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 这里检测token是否已经退出，设置为无效
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token 无效或者 redis 有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
	}
}
