package mq

import (
	"context"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

// Producer is a Redpanda/Kafka producer that publishes messages to topics.
// All publishes are fire-and-forget — errors are logged but never block the API.
// When disabled (no brokers configured), the fallback direct ClickHouse path is used.
type Producer struct {
	writers map[string]*kafka.Writer
	brokers []string
	enabled bool
}

// NewProducer creates a connected Redpanda producer.
// If brokers is empty, the producer is disabled (direct CH writes used instead).
func NewProducer(brokers []string) (*Producer, error) {
	if len(brokers) == 0 {
		return &Producer{enabled: false}, nil
	}

	writers := map[string]*kafka.Writer{
		TopicAgentMetrics: newWriter(brokers, TopicAgentMetrics),
		TopicProbeResults: newWriter(brokers, TopicProbeResults),
	}

	// Verify connectivity with a short ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := kafka.DialContext(ctx, "tcp", brokers[0])
	if err != nil {
		for _, w := range writers {
			w.Close()
		}
		return nil, err
	}
	conn.Close()

	log.Printf("Redpanda producer connected to %v", brokers)
	return &Producer{writers: writers, brokers: brokers, enabled: true}, nil
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.Hash{},   // partition by key (host_id)
		BatchTimeout: 5 * time.Millisecond,
		BatchSize:    100,
		Async:        true, // fire-and-forget; errors via Logger
		Logger:       kafka.LoggerFunc(func(msg string, args ...interface{}) {}),
		ErrorLogger:  kafka.LoggerFunc(func(msg string, args ...interface{}) {
			log.Printf("kafka error topic=%s: "+msg, append([]interface{}{topic}, args...)...)
		}),
		RequiredAcks: kafka.RequireOne,
		Compression:  kafka.Snappy,
	}
}

func (p *Producer) Enabled() bool { return p.enabled }

// Publish sends a message to the given topic asynchronously.
// key is used for partitioning (typically host_id for ordering per host).
func (p *Producer) Publish(ctx context.Context, topic string, key, value []byte) {
	if !p.enabled {
		return
	}
	w, ok := p.writers[topic]
	if !ok {
		return
	}
	if err := w.WriteMessages(ctx, kafka.Message{Key: key, Value: value}); err != nil {
		log.Printf("mq publish error topic=%s: %v", topic, err)
	}
}

func (p *Producer) Close() {
	if !p.enabled {
		return
	}
	for _, w := range p.writers {
		_ = w.Close()
	}
}
