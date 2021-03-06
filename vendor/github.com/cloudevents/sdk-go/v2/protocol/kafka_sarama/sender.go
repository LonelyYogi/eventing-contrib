package kafka_sarama

import (
	"context"

	"github.com/Shopify/sarama"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// Sender implements binding.Sender that sends messages to a specific receiverTopic using sarama.SyncProducer
type Sender struct {
	topic        string
	syncProducer sarama.SyncProducer
}

// NewSender returns a binding.Sender that sends messages to a specific receiverTopic using sarama.SyncProducer
func NewSender(brokers []string, saramaConfig *sarama.Config, topic string, options ...SenderOptionFunc) (*Sender, error) {
	// Force this setting because it's required by sarama SyncProducer
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, saramaConfig)
	if err != nil {
		return nil, err
	}

	return makeSender(producer, topic, options...), nil
}

// NewSenderFromClient returns a binding.Sender that sends messages to a specific receiverTopic using sarama.SyncProducer
func NewSenderFromClient(client sarama.Client, topic string, options ...SenderOptionFunc) (*Sender, error) {
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, err
	}

	return makeSender(producer, topic, options...), nil
}

func makeSender(syncProducer sarama.SyncProducer, topic string, options ...SenderOptionFunc) *Sender {
	s := &Sender{
		topic:        topic,
		syncProducer: syncProducer,
	}
	for _, o := range options {
		o(s)
	}
	return s
}

func (s *Sender) Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error {
	var err error
	defer m.Finish(err)

	kafkaMessage := sarama.ProducerMessage{Topic: s.topic}

	if err = WriteProducerMessage(ctx, m, &kafkaMessage, transformers...); err != nil {
		return err
	}

	_, _, err = s.syncProducer.SendMessage(&kafkaMessage)
	// Somebody closed the client while sending the message, so no problem here
	if err == sarama.ErrClosedClient {
		return nil
	}
	return err
}

func (s *Sender) Close(ctx context.Context) error {
	// If the Sender was built with NewSenderFromClient, this Close will close only the producer,
	// otherwise it will close the whole client
	return s.syncProducer.Close()
}
