package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/interactive/repository"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/logger"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/pkg/saramax"
	"time"
)

type InteractiveReadEventConsumer struct {
	client sarama.Client
	repo   repository.InteractiveRepository
	l      logger.Logger
}

func NewInteractiveReadEventConsumer(client sarama.Client, repo repository.InteractiveRepository, l logger.Logger) *InteractiveReadEventConsumer {
	return &InteractiveReadEventConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (i *InteractiveReadEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive", i.client)
	if err != nil {
		return err
	}

	// 这里的go routine实际上是一个无限循环函数，只要kafka一直在运行，这里就一直在
	go func() {
		err = cg.Consume(context.Background(), []string{"read_article"},
			saramax.NewHandler[ReadEvent](i.l, i.Consume))
		if err != nil {
			i.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()

	return err

}

// Consume 这个不是幂等的
// 处理真正的业务逻辑
func (i *InteractiveReadEventConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	_, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return i.repo.IncrReadCnt(context.Background(), "article", t.Aid)

}
