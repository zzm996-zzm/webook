package startup

import (
	"webook/internal/service/oauth2/wechat"
)

func InitWechatService() wechat.Service {
	return wechat.NewService("", "")
}
