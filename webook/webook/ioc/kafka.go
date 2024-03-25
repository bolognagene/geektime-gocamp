package ioc

import (
	"github.com/IBM/sarama"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	type Config struct {
		Addr []string `yaml:"addrs"`
	}

	saramaCfg := sarama.NewConfig()
	// a backwards incompatible change
	saramaCfg.Producer.Return.Successes = true
	var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addr, saramaCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}
	return producer
}

// NewConsumers 面临的问题依旧是所有的 Consumer 在这里注册一下
func NewConsumers(c1 *article.InteractiveReadEventBatchConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
