package ioc

import (
	"context"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func InitMongoDB() *mongo.Client {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context,
			startedEvent *event.CommandStartedEvent) {
			fmt.Println(startedEvent.Command)
		},
		Succeeded: func(ctx context.Context,
			succeededEvent *event.CommandSucceededEvent) {
			fmt.Println(succeededEvent.CommandName, succeededEvent.DurationNanos)
		},
		Failed: func(ctx context.Context,
			failedEvent *event.CommandFailedEvent) {
			fmt.Println(failedEvent.CommandName, failedEvent.DurationNanos)
		},
	}
	addr := viper.GetString("mongodb.addr")
	if addr == "" {
		addr = "mongodb://root:example@localhost:27017/"
	}
	opts := options.Client().
		ApplyURI(addr).
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		panic(err)
	}

	return client
}

func InitSnowflakeNode() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	return node
}
