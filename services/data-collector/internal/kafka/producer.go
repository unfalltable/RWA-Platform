package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	writers map[string]*kafka.Writer
	logger  *logrus.Logger
}

func NewProducer(brokers []string) (*Producer, error) {
	producer := &Producer{
		writers: make(map[string]*kafka.Writer),
		logger:  logrus.New(),
	}

	// 预创建常用topic的writer
	topics := []string{
		"price-updates",
		"blockchain-events", 
		"token-transfers",
		"news-updates",
		"risk-alerts",
		"system-events",
	}

	for _, topic := range topics {
		writer := &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    100,
			Async:        true,
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Snappy,
		}
		producer.writers[topic] = writer
	}

	return producer, nil
}

func (p *Producer) PublishMessage(topic string, key string, message interface{}) error {
	writer, exists := p.writers[topic]
	if !exists {
		// 动态创建writer
		writer = &kafka.Writer{
			Addr:         p.writers["price-updates"].Addr,
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			BatchTimeout: 10 * time.Millisecond,
			BatchSize:    100,
			Async:        true,
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Snappy,
		}
		p.writers[topic] = writer
	}

	// 序列化消息
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 发送消息
	kafkaMessage := kafka.Message{
		Key:   []byte(key),
		Value: messageBytes,
		Time:  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := writer.WriteMessages(ctx, kafkaMessage); err != nil {
		p.logger.Errorf("Failed to write message to topic %s: %v", topic, err)
		return err
	}

	p.logger.Debugf("Published message to topic %s with key %s", topic, key)
	return nil
}

func (p *Producer) PublishBatch(topic string, messages []kafka.Message) error {
	writer, exists := p.writers[topic]
	if !exists {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := writer.WriteMessages(ctx, messages...); err != nil {
		p.logger.Errorf("Failed to write batch messages to topic %s: %v", topic, err)
		return err
	}

	p.logger.Debugf("Published %d messages to topic %s", len(messages), topic)
	return nil
}

func (p *Producer) Close() error {
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil {
			p.logger.Errorf("Failed to close writer for topic %s: %v", topic, err)
		}
	}
	return nil
}

// Consumer 消费者结构
type Consumer struct {
	reader *kafka.Reader
	logger *logrus.Logger
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader: reader,
		logger: logrus.New(),
	}
}

func (c *Consumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	HandleMessage(ctx context.Context, message kafka.Message) error
}

// StartConsumer 启动消费者
func (c *Consumer) StartConsumer(ctx context.Context, handler MessageHandler) {
	c.logger.Infof("Starting consumer for topic: %s", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopped")
			return
		default:
			message, err := c.ReadMessage(ctx)
			if err != nil {
				c.logger.Errorf("Failed to read message: %v", err)
				continue
			}

			if err := handler.HandleMessage(ctx, message); err != nil {
				c.logger.Errorf("Failed to handle message: %v", err)
				continue
			}

			c.logger.Debugf("Processed message from topic %s, offset %d", 
				message.Topic, message.Offset)
		}
	}
}
