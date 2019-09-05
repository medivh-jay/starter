package log

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
)

var (
	ErrCannotCreateIndex = fmt.Errorf("cannot create index")
)

type IndexNameFunc func() string

type fireFunc func(entry *logrus.Entry, hook *ElasticHook) error

type ElasticHook struct {
	client    *elastic.Client
	host      string
	index     IndexNameFunc
	levels    []logrus.Level
	ctx       context.Context
	ctxCancel context.CancelFunc
	fireFunc  fireFunc
}

type message struct {
	Host      string
	Timestamp string `json:"@timestamp"`
	Message   string
	Data      logrus.Fields
	Level     string
}

func NewElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewElasticHookWithFunc(client, host, level, func() string { return index })
}

func NewAsyncElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewAsyncElasticHookWithFunc(client, host, level, func() string { return index })
}

func NewBulkProcessorElasticHook(client *elastic.Client, host string, level logrus.Level, index string) (*ElasticHook, error) {
	return NewBulkProcessorElasticHookWithFunc(client, host, level, func() string { return index })
}

func NewElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, syncFireFunc)
}

func NewAsyncElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	return newHookFuncAndFireFunc(client, host, level, indexFunc, asyncFireFunc)
}

func NewBulkProcessorElasticHookWithFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc) (*ElasticHook, error) {
	fireFunc, err := makeBulkFireFunc(client)
	if err != nil {
		return nil, err
	}
	return newHookFuncAndFireFunc(client, host, level, indexFunc, fireFunc)
}

func newHookFuncAndFireFunc(client *elastic.Client, host string, level logrus.Level, indexFunc IndexNameFunc, fireFunc fireFunc) (*ElasticHook, error) {
	var levels []logrus.Level
	for _, l := range []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
		logrus.TraceLevel,
	} {
		if l <= level {
			levels = append(levels, l)
		}
	}

	ctx, cancel := context.WithCancel(context.TODO())

	exists, err := client.IndexExists(indexFunc()).Do(ctx)
	if err != nil {
		cancel()
		return nil, err
	}
	if !exists {
		createIndex, err := client.CreateIndex(indexFunc()).Do(ctx)
		if err != nil {
			cancel()
			return nil, err
		}
		if !createIndex.Acknowledged {
			cancel()
			return nil, ErrCannotCreateIndex
		}
	}

	return &ElasticHook{
		client:    client,
		host:      host,
		index:     indexFunc,
		levels:    levels,
		ctx:       ctx,
		ctxCancel: cancel,
		fireFunc:  fireFunc,
	}, nil
}

func (hook *ElasticHook) Fire(entry *logrus.Entry) error {
	return func(hook *ElasticHook, entry *logrus.Entry) error {
		return hook.fireFunc(entry, hook)
	}(hook, entry)
}

func asyncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	go syncFireFunc(entry, hook)
	return nil
}

func createMessage(entry *logrus.Entry, hook *ElasticHook) *message {
	level := entry.Level.String()

	if e, ok := entry.Data[logrus.ErrorKey]; ok && e != nil {
		if err, ok := e.(error); ok {
			entry.Data[logrus.ErrorKey] = err.Error()
		}
	}

	return &message{
		hook.host,
		entry.Time.UTC().Format(time.RFC3339Nano),
		entry.Message,
		entry.Data,
		strings.ToUpper(level),
	}
}

func syncFireFunc(entry *logrus.Entry, hook *ElasticHook) error {
	_, err := hook.client.
		Index().
		Index(hook.index()).
		BodyJson(*createMessage(entry, hook)).
		Do(hook.ctx)
	return err
}

func makeBulkFireFunc(client *elastic.Client) (fireFunc, error) {
	processor, err := client.BulkProcessor().
		Name("elastic.log.bulk.processor").
		Workers(3).
		BulkActions(-1).
		BulkSize(-1).
		FlushInterval(10 * time.Second).
		Do(context.Background())

	return func(entry *logrus.Entry, hook *ElasticHook) error {
		r := elastic.NewBulkIndexRequest().Index(hook.index()).Doc(*createMessage(entry, hook))
		processor.Add(r)
		return nil
	}, err
}

func (hook *ElasticHook) Levels() []logrus.Level {
	return hook.levels
}

func (hook *ElasticHook) Cancel() {
	hook.ctxCancel()
}
