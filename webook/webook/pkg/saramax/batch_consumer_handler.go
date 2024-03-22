package saramax

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"time"
)

type BatchHandler[T any] struct {
	l  logger.Logger
	fn func(msgs []*sarama.ConsumerMessage, ts []T) error
	// 用 option 模式来设置这个 batchSize 和 duration
	batchSize     int
	batchDuration time.Duration
}

func NewBatchHandler[T any](l logger.Logger, fn func(msgs []*sarama.ConsumerMessage, ts []T) error) *BatchHandler[T] {
	return &BatchHandler[T]{
		l:             l,
		fn:            fn,
		batchSize:     10,
		batchDuration: time.Second,
	}
}

func (h BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgsCh := claim.Messages()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), h.batchDuration)
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, h.batchSize)
		ts := make([]T, 0, h.batchSize)

		for i := 0; i < h.batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgsCh:
				if !ok {
					cancel()
					// 代表消费者被关闭了
					return nil
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					h.l.Error("反序列化消息失败",
						logger.Error(err),
						logger.String("topic", msg.Topic),
						logger.Int64("partition", int64(msg.Partition)),
						logger.Int64("offset", msg.Offset))
					continue
				}
				msgs = append(msgs, msg)
				ts = append(ts, t)
			}
		}
		cancel()

		if len(msgs) == 0 {
			continue
		}
		err := h.fn(msgs, ts)
		if err != nil {
			h.l.Error("调用业务批量接口失败",
				logger.Error(err))
			// 你这里整个批次都要记下来

			// 还要继续往前消费
		}
		for _, msg := range msgs {
			// 这样，万无一失
			session.MarkMessage(msg, "")
		}

	}

}
