package ioc

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"webook/internal/repository/dao"
)

func InitMongoDB() *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Println(evt.Command)
		},
	}
	opts := options.Client().
		ApplyURI("mongodb://root:example@localhost:27017/").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	mdb := client.Database("webook")

	err = dao.InitCollection(mdb)
	if err != nil {
		panic(err)
	}

	return mdb
}
