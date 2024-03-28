package article

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/saramax"
	"time"
)

type InteractiveReadEventBatchConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventBatchConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventBatchConsumer {
	return &InteractiveReadEventBatchConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (i *InteractiveReadEventBatchConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	// 这里的go routine实际上是一个无限循环函数，只要kafka一直在运行，这里就一直在
	go func() {
		err = cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewBatchHandler[ReadEvent](i.l, i.Consume))
		if err != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()

	return err

}

// Consume 这个不是幂等的
// 处理真正的业务逻辑
func (i *InteractiveReadEventBatchConsumer) Consume(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := i.repo.BatchIncrReadCnt(ctx, "article", ids)
	if err != nil {
		i.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil

}

// ConsumeV1 这个不是幂等的
// 处理真正的业务逻辑
// 将read_cnt计数只添加到redis里
func (i *InteractiveReadEventBatchConsumer) ConsumeV1(msgs []*sarama.ConsumerMessage, ts []ReadEvent) error {
	ids := make([]int64, 0, len(ts))
	for _, evt := range ts {
		ids = append(ids, evt.Aid)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := i.repo.BatchIncrReadCnt(ctx, "article", ids)
	if err != nil {
		i.l.Error("批量增加阅读计数失败",
			logger.Field{Key: "ids", Value: ids},
			logger.Error(err))
	}
	return nil

}
