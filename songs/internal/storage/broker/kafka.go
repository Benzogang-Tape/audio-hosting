package broker

import (
	"context"
	"net"
	"strconv"

	"dev.gaijin.team/go/golib/e"
	"dev.gaijin.team/go/golib/fields"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	w *kafka.Writer
}

func Connect(topic string, brokers []string) (*KafkaProducer, error) {
	err := createTopic(topic, brokers)
	if err != nil {
		return nil, e.NewFrom("creating topic", err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{ //nolint:exhaustruct
		Brokers: brokers,
		Topic:   topic,
	})

	return &KafkaProducer{
		w: writer,
	}, nil
}

func createTopic(topic string, brokers []string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return e.NewFrom("connecting to kafka", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return e.NewFrom("getting controller", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return e.NewFrom("connecting to controller", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{{ //nolint:exhaustruct
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return e.NewFrom("creating topic", err, fields.F("topic", topic))
	}

	return nil
}

func (k *KafkaProducer) SendReleasedMessages(ctx context.Context, messages []SongReleasedMessage) error {
	msgs := make([]kafka.Message, len(messages))
	for i := range messages {
		msgs[i] = kafka.Message{ //nolint:exhaustruct
			Value: messages[i].Bytes(),
		}
	}

	err := k.w.WriteMessages(ctx, msgs...)
	if err != nil {
		return e.NewFrom("sending messages to kafka", err)
	}

	return nil
}

func (k *KafkaProducer) Close() error {
	err := k.w.Close()
	if err != nil {
		return e.NewFrom("disconnecting from kafka", err)
	}

	return nil
}
