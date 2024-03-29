package redisx

import (
	"context"
	"fmt"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/key_expired_event"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	client redis.Cmdable
	evts   []key_expired_event.KeyExpiredEvent
}

func NewHandler(client redis.Cmdable, evts []key_expired_event.KeyExpiredEvent) *Handler {
	return &Handler{
		client: client,
		evts:   evts,
	}
}

func (h *Handler) NotifyKeyExpiredEvent() {
	cli, ok := h.client.(*redis.Client)
	if ok {
		_, err := cli.Do(context.Background(), "CONFIG", "SET", "notify-keyspace-events", "Ex").Result()
		if err != nil {
			panic(err)
		}
		pubSub := cli.PSubscribe(context.Background(), "__keyevent@0__:expired")

		go func() {
			// 创建一个接收通道以接收订阅的消息
			channel := pubSub.Channel()
			// 开始监听订阅的消息
			for msg := range channel {
				fmt.Printf("Received expired key message: %s\n", msg.Payload)

				for _, evt := range h.evts {
					evt.Process(msg.Payload)
				}
			}
		}()

	} else {
		panic("notify redis expired key initialize failed!")
	}
}
