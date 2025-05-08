package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

const KafkaBrokerAddress = "kafka:9092"

func InitKafka(kafkaTopicName string) (sarama.Client, error) {
	kafkaNumPartitions := int32(1)
	kafkaReplicationFactor := int16(1)
	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0
	config.Producer.Return.Successes = true

	kafkaClient, err := sarama.NewClient([]string{KafkaBrokerAddress}, config)
	if err != nil {
		slog.Error("Error creating Kafka client", "error", err)
		return nil, fmt.Errorf("InitKafka error creating Kafka client: %w", err)
	}
	err = createTopicIfNotExists(
		kafkaTopicName,
		kafkaNumPartitions,
		kafkaReplicationFactor,
	)
	if err != nil {
		kafkaClient.Close()
		slog.Error("Error ensuring Kafka topic exists", "topic", kafkaTopicName, "error", err)
		return nil, fmt.Errorf("error ensuring Kafka topic exists: %w", err)
	}

	return kafkaClient, nil
}

func createTopicIfNotExists(
	topic string,
	numPartitions int32,
	replicationFactor int16,
) error {
	broker := sarama.NewBroker(KafkaBrokerAddress)

	config := sarama.NewConfig()
	config.Version = sarama.V4_0_0_0

	err := broker.Open(config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka broker: %w", err)
	}
	defer broker.Close()

	request := sarama.NewCreateTopicsRequest(config.Version, map[string]*sarama.TopicDetail{
		KafkaBrokerAddress: {
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}, time.Minute)

	_, err = broker.CreateTopics(request)
	if err != nil {
		return fmt.Errorf("failed to create Kafka topic '%s': %w", topic, err)
	}

	return nil
}

type Consumer struct {
	ready   chan bool
	consume func(value []byte)
}

func NewConsumer(client sarama.Client, consume func(value []byte), topics []string) {
	cg, err := sarama.NewConsumerGroupFromClient("consumers", client)
	if err != nil {
		slog.Error("NewConsumer can't create consumer group", "error", err.Error())
		os.Exit(1)
	}

	consumer := Consumer{
		ready:   make(chan bool),
		consume: consume,
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := cg.Consume(ctx, topics, &consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				slog.Error("Error from consumer", "error", err.Error())
				os.Exit(1)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready
	slog.Info("Consumer up and running", "topics", topics)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			slog.Info("terminating context cancelled")
			keepRunning = false
		case <-sigterm:
			slog.Info("terminating via signal")
			keepRunning = false
		}
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		slog.Error("Error closing client", "error", err)
		os.Exit(1)
	}
}

// Taken from examples https://github.com/IBM/sarama/blob/v1.45.1/examples/consumergroup/main.go

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				slog.Info("message channel was closed")
				return nil
			}
			slog.Info(
				"Message claimed",
				"timestamp", message.Timestamp,
				"topic", message.Topic,
			)
			consumer.consume(message.Value)
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
