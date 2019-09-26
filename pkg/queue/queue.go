package queue

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
	"starter/pkg/app"
	"sync"
)

type (
	// Consumer 消费者
	Consumer struct {
		conf      kafka.ConfigMap
		once      sync.Once
		topics    []string // 消费标签列表
		reBalance kafka.RebalanceCb
		consumer  *kafka.Consumer
		stop      bool // 是否关闭当前消费者
	}
	// Producer 生产者
	Producer struct {
		conf      kafka.ConfigMap
		once      sync.Once
		connected sync.Once
		topics    []string // 消费标签列表
		reBalance kafka.RebalanceCb
		producer  *kafka.Producer
		stop      chan bool // 是否关闭当前消费者
	}
)

// SetTopics 设置监听的标签
func (q *Consumer) SetTopics(topics ...string) *Consumer {
	q.topics = topics
	return q
}

// SetReBalance 自定义回调方法
func (q *Consumer) SetReBalance(reBalanceCb kafka.RebalanceCb) *Consumer {
	q.reBalance = reBalanceCb
	return q
}

// Stop 关闭
func (q *Consumer) Stop() {
	q.stop = true
}

// SetConfig 设置kafka配置
//  配置参数详阅 https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
func (q *Consumer) SetConfig(key string, value interface{}) *Consumer {
	_ = q.conf.SetKey(key, value)
	return q
}

// Do 将会开启一个协程运行消费者程序
func (q *Consumer) Do(consumer func(consumer *kafka.Consumer, message *kafka.Message)) *Consumer {
	if len(q.topics) == 0 {
		app.Logger().WithField("log_type", "pkg.queue.queue").Error("Consumer: topics is empty")
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
			}
			msg, err := q.consumer.ReadMessage(-1)
			if err == nil {
				consumer(q.consumer, msg)
			} else {
				app.Logger().WithField("log_type", "pkg.queue.queue").WithField("kafka_msg", msg).Error(err)
			}
		}
	}()
	return q
}

// NewConsumer 获得一个新的消费者
func NewConsumer() *Consumer {
	var queue = new(Consumer)
	queue.once.Do(func() {
		_ = app.Config().Bind("application", "kafka", &queue.conf)
	})

	return queue
}

// SetConfig 设置kafka配置
//  配置参数详阅 https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
func (producer *Producer) SetConfig(key string, value interface{}) *Producer {
	if producer.conf == nil {
		producer.conf = make(kafka.ConfigMap)
	}
	_ = producer.conf.SetKey(key, value)
	return producer
}

func (producer *Producer) handle() {
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

// Send 发送消息, value如果类型不是 *kafka.Message, 将被转为默认的 json []byte
func (producer *Producer) Send(topic string, value interface{}) error {
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
		var data []byte
		data, err = jsoniter.Marshal(value)
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

// Flush 刷新并等待未完成的操作
//  see https://godoc.org/gopkg.in/confluentinc/confluent-kafka-go.v1/kafka#Producer.Flush
func (producer *Producer) Flush() {
	producer.producer.Flush(15 * 1000)
}

// Stop 关闭生产者发送事件监听协程,并断开与kafka的连接
func (producer *Producer) Stop() {
	producer.stop <- true
	producer.producer.Close()
}

// NewProducer 获得一个新的生产者,生产者不需要每次使用都重新获取, 直接保存一个全局变量就行
func NewProducer() *Producer {
	var producer = new(Producer)
	producer.once.Do(func() {
		var conf kafka.ConfigMap
		_ = app.Config().Bind("application", "kafka", &conf)
		if val, ok := conf["bootstrap.servers"]; ok {
			producer.SetConfig("bootstrap.servers", val)
		}
	})

	return producer
}
