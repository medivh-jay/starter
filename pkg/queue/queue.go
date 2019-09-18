package queue

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"starter/pkg/app"
	"sync"
)

type (
	// 消费者
	consumer struct {
		conf      kafka.ConfigMap
		once      sync.Once
		topics    []string // 消费标签列表
		reBalance kafka.RebalanceCb
		consumer  *kafka.Consumer
		stop      bool // 是否关闭当前消费者
	}
	// 生产者
	producer struct {
		conf      kafka.ConfigMap
		once      sync.Once
		connected sync.Once
		topics    []string // 消费标签列表
		reBalance kafka.RebalanceCb
		producer  *kafka.Producer
		stop      chan bool // 是否关闭当前消费者
	}
)

func (q *consumer) SetTopics(topics ...string) *consumer {
	q.topics = topics
	return q
}

func (q *consumer) SetReBalance(reBalanceCb kafka.RebalanceCb) *consumer {
	q.reBalance = reBalanceCb
	return q
}

func (q *consumer) Stop() {
	q.stop = true
}

// 设置kafka配置
//  配置参数详阅 https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
func (q *consumer) SetConfig(key string, value interface{}) *consumer {
	_ = q.conf.SetKey(key, value)
	return q
}

// 将会开启一个协程运行消费者程序
func (q *consumer) Do(consumer func(consumer *kafka.Consumer, message *kafka.Message)) *consumer {
	if len(q.topics) == 0 {
		app.Logger().WithField("log_type", "pkg.queue.queue").Error("consumer: topics is empty")
		return q
	}

	client, err := kafka.NewConsumer(&q.conf)
	if err != nil {
		app.Logger().WithField("log_type", "pkg.queue.queue").Error(err)
	}
	q.consumer = client

	_ = q.consumer.SubscribeTopics(q.topics, q.reBalance)
	go func() {
		for {
			if q.stop {
				_ = q.consumer.Close()
				return
			} else {
				msg, err := q.consumer.ReadMessage(-1)
				if err == nil {
					consumer(q.consumer, msg)
				} else {
					app.Logger().WithField("log_type", "pkg.queue.queue").WithField("kafka_msg", msg).Error(err)
				}
			}
		}
	}()
	return q
}

// 获得一个新的消费者
func NewConsumer() *consumer {
	var queue = new(consumer)
	queue.once.Do(func() {
		_ = app.Config().Bind("application", "kafka", &queue.conf)
	})

	return queue
}

// 设置kafka配置
//  配置参数详阅 https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
func (producer *producer) SetConfig(key string, value interface{}) *producer {
	if producer.conf == nil {
		producer.conf = make(kafka.ConfigMap)
	}
	_ = producer.conf.SetKey(key, value)
	return producer
}

func (producer *producer) handle() {
	go func() {
		for {
			select {
			case <-producer.stop:
				return
			case event := <-producer.producer.Events():
				switch ev := event.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						app.Logger().WithField("log_type", "pkg.queue.queue").WithField("kafka_event", ev).Error("kafka: delivery failed")
					} else {
						if gin.IsDebugging() {
							app.Logger().WithField("log_type", "pkg.queue.queue").Debug(ev)
						}
					}
				}
			}
		}
	}()
}

// 发送消息, value如果类型不是 *kafka.Message, 将被转为默认的 json []byte
func (producer *producer) Send(topic string, value interface{}) error {
	producer.connected.Do(func() {
		var err error
		producer.producer, err = kafka.NewProducer(&producer.conf)
		if err != nil {
			app.Logger().WithField("log_type", "pkg.queue.queue").Error(err)
		} else {
			producer.handle()
		}
	})
	var err error
	switch value.(type) {
	case *kafka.Message:
		err = producer.producer.Produce(value.(*kafka.Message), nil)
	default:
		data, err := jsoniter.Marshal(value)
		if err != nil {
			app.Logger().WithField("log_type", "pkg.queue.queue").Error(err)
		}
		err = producer.producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}, nil)
	}
	return err
}

func (producer *producer) Flush() {
	producer.producer.Flush(15 * 1000)
}

// 关闭生产者发送事件监听协程,并断开与kafka的连接
func (producer *producer) Stop() {
	producer.stop <- true
	producer.producer.Close()
}

// 获得一个新的生产者,生产者不需要每次使用都重新获取, 直接保存一个全局变量就行
func NewProducer() *producer {
	var producer = new(producer)
	producer.once.Do(func() {
		var conf kafka.ConfigMap
		_ = app.Config().Bind("application", "kafka", &conf)
		if val, ok := conf["bootstrap.servers"]; ok {
			producer.SetConfig("bootstrap.servers", val)
		}
	})

	return producer
}
