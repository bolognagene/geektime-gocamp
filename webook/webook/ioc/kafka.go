package ioc

import (
	"github.com/IBM/sarama"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events"
	"github.com/bolognagene/geektime-gocamp/geektime-gocamp/webook/webook/internal/events/article"
	"github.com/spf13/viper"
)

func InitKafka() sarama.Client {
	/*type Config struct {
		Addr []string `yaml:"addrs"`
	}*/

	saramaCfg := sarama.NewConfig()
	// a backwards incompatible change
	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false //手动提交
	/*var cfg Config
	err := viper.UnmarshalKey("kafka", &cfg)
	if err != nil {
		panic(err)
	}
	client, err := sarama.NewClient(cfg.Addr, saramaCfg)*/
	addrs := []string{viper.GetString("kafka.addrs")}
	if len(addrs) == 0 {
		addrs = []string{"localhost:9094"}
	}
	client, err := sarama.NewClient(addrs, saramaCfg)
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
/*func NewConsumers(c1 *article.InteractiveReadEventBatchConsumer) []events.Consumer {
	return []events.Consumer{c1}
}*/
func NewConsumers(c1 *article.InteractiveReadEventConsumer) []events.Consumer {
	return []events.Consumer{c1}
}
