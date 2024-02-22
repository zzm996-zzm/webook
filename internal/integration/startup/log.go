package startup

import "webook/pkg/logger"

func InitLogger() logger.Logger {
	return logger.NewNopLogger()
}
