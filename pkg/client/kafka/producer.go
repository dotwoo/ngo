// Copyright Ngo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kafka

import (
	"errors"
	"sync"
	"time"

	"github.com/NetEase-Media/ngo/pkg/adapter/log"
	"github.com/Shopify/sarama"
)

type ProducerMessage struct {
	Topic string
	Key   string
	Value string
}

type RecordMetadata struct {
	Topic     string
	KeySize   int
	ValueSize int
	Offset    int64
	Partition int32
}

type Producer struct {
	client  sarama.AsyncProducer
	opt     Options
	logger  *log.NgoLogger
	runChan chan struct{}
	wg      sync.WaitGroup
}

type Callback func(*RecordMetadata, error)

func (p *Producer) Options() Options {
	return p.opt
}

// Send 是异步发送接口
func (p *Producer) Send(topic, value string, cb Callback) {
	p.SendMessage(ProducerMessage{Topic: topic, Value: value}, cb)
}

// SendMessage 是异步发送接口
func (p *Producer) SendMessage(message ProducerMessage, cb Callback) {
	meta := newMetaData()
	meta.cb = cb
	m := &sarama.ProducerMessage{
		Topic:    message.Topic,
		Key:      nil,
		Value:    sarama.StringEncoder(message.Value),
		Metadata: meta,
	}
	if len(message.Key) != 0 {
		m.Key = sarama.StringEncoder(message.Key)
	}
	p.client.Input() <- m
}

// SyncSend 是同步发送接口。
func (p *Producer) SyncSend(topic, value string) error {
	return p.SyncSendMessage(ProducerMessage{Topic: topic, Value: value})
}

// SyncSendMessage 是同步发送接口。
func (p *Producer) SyncSendMessage(message ProducerMessage) error {
	meta := newMetaData()
	meta.resChan = make(chan error)
	m := &sarama.ProducerMessage{
		Topic:    message.Topic,
		Key:      nil,
		Value:    sarama.StringEncoder(message.Value),
		Metadata: meta,
	}
	if len(message.Key) != 0 {
		m.Key = sarama.StringEncoder(message.Key)
	}

	p.client.Input() <- m
	select {
	case err := <-meta.resChan:
		return err
	case <-time.NewTimer(time.Second * 10).C:
		// 防止异常卡死
		return errors.New("send timeout")
	}
}

// run 启动后台任务，接收结果和错误
func (p *Producer) run() {
	p.wg.Add(1)
	go p.receiveSuccess()

	p.wg.Add(1)
	go p.receiveError()
}

// receiveSuccess 接收成功回复，记录结果
func (p *Producer) receiveSuccess() {
	defer p.wg.Done()
	for {
		select {
		case s, ok := <-p.client.Successes():
			if !ok {
				return
			}
			p.handle(s, nil)
		}
	}
}

// receiveError 接收错误回复
func (p *Producer) receiveError() {
	defer p.wg.Done()
	for {
		select {
		case e, ok := <-p.client.Errors():
			if !ok {
				return
			}
			p.handle(e.Msg, e.Err)
		}
	}
}

// handle 处理异步消息的发送结果
func (p *Producer) handle(msg *sarama.ProducerMessage, err error) {
	log.Tracef("receive send response %+v, error %v", msg, err)
	meta := msg.Metadata.(*metaData)
	if meta.resChan != nil {
		meta.resChan <- err
		close(meta.resChan)
	}

	var ks, vs int
	if msg.Key != nil {
		ks = msg.Key.Length()
	}
	if msg.Value != nil {
		vs = msg.Value.Length()
	}

	if meta.cb != nil {
		meta.cb(&RecordMetadata{
			Topic:     msg.Topic,
			KeySize:   ks,
			ValueSize: vs,
			Offset:    msg.Offset,
			Partition: msg.Partition,
		}, err)
	}
}

// Close 关闭客户端，等待缓冲区完成读写再返回
func (p *Producer) Close() {
	p.client.AsyncClose()
	p.wg.Wait()
}

// NewProducer 创建一个异步的生产者
func NewProducer(opt *Options) (*Producer, error) {
	config, err := newProducerConfig(opt)
	if err != nil {
		return nil, err
	}

	p, err := sarama.NewAsyncProducer(opt.Addr, config)
	if err != nil {
		return nil, err
	}

	producer := &Producer{
		client:  p,
		logger:  log.WithField("kafka", opt.Name),
		opt:     *opt,
		runChan: make(chan struct{}),
	}
	producer.run()
	return producer, nil
}

func newProducerConfig(opt *Options) (*sarama.Config, error) {
	config := sarama.NewConfig()
	version, err := sarama.ParseKafkaVersion(opt.Version)
	if err != nil {
		return nil, err
	}
	config.Version = version
	config.ChannelBufferSize = 1024

	config.Net.MaxOpenRequests = opt.MaxOpenRequests
	config.Net.DialTimeout = opt.DialTimeout
	config.Net.ReadTimeout = opt.ReadTimeout
	config.Net.WriteTimeout = opt.WriteTimeout

	config.Metadata.Retry.Max = opt.Metadata.Retries
	config.Metadata.Timeout = opt.Metadata.Timeout

	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = opt.Producer.Acks
	config.Producer.Timeout = opt.Producer.Timeout
	config.Producer.Retry.Max = opt.Producer.Retries
	config.Producer.Flush.Bytes = opt.Producer.MaxFlushBytes
	config.Producer.Flush.Messages = opt.Producer.MaxFlushMessages
	config.Producer.Flush.Frequency = opt.Producer.FlushFrequency
	config.Producer.Idempotent = opt.Producer.Idempotent

	return config, nil
}

func newMetaData() *metaData {
	return &metaData{
		startTime: time.Now(),
	}
}

// metaData 注册到message中，主要用来监控
type metaData struct {
	startTime time.Time
	resChan   chan error // 回复channel，可以将异步调用变成同步
	cb        Callback
}
