package main

import (
	"webook/internal/events"
	"webook/pkg/saramax"

	"github.com/gin-gonic/gin"
)

type App struct {
	server       *gin.Engine
	consumers    []events.Consumer
	kafkaMonitor *saramax.MonitorMessage
}
