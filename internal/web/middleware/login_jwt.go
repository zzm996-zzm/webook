package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
	"time"
	"webook/internal/web"
)

type LoginJWTMiddlewareBuilder struct {
}

func (*LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 不需要校验
		if ctx.Request.URL.Path == "/users/signup" ||
			ctx.Request.URL.Path == "/users/login" {
			return
		}

		authCode := ctx.GetHeader("Authorization")
		if authCode == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		segs := strings.Split(authCode, " ")
		if len(segs) != 2 {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		tokenStr := segs[1]
		var uc = web.UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return web.JWTKEY, nil
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

		// 剩余过期时间 < 50s 就要刷新
		expireTime := uc.ExpiresAt
		if expireTime.Sub(time.Now()) < time.Second*50 {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
			newToken, err := token.SigningString()
			if err != nil {
				fmt.Println("token刷新失败")
			} else {
				ctx.Header("x-jwt-token", newToken)
			}
		}

		ctx.Set("user", uc)
	}
}
