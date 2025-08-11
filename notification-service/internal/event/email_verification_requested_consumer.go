package event

import (
	"sync"

	"github.com/raphaeldiscky/go-micro-template/pkg/mq"
)

type EmailVerificationRequestedEvent struct {
	Token string `json:"token"`
}

type EmailVerificationRequestedConsumer struct {
	Consumer *mq.KafkaConsumer
	topic    string
	wg       *sync.WaitGroup
}

func NewEmailVerificationRequestedConsumer(consumer *mq.KafkaConsumer, topic string) *EmailVerificationRequestedConsumer {
	return &EmailVerificationRequestedConsumer{
		Consumer: consumer,
		topic:    topic,
		wg:       &sync.WaitGroup{},
	}
}

// // Consume starts the Kafka consumer for the email verification requested events.
// func (c *EmailVerificationRequestedConsumer) Consume(ctx context.Context) error {
// 	c.wg.Add(1)
// 	c.Consumer.Start()
// }

// // Handler for email verification requested events.
// func (c *EmailVerificationRequestedConsumer) Handler() error {
// 	var event EmailVerificationRequestedEvent
// 	if err := msg.Unmarshal(&event); err != nil {
// 		return err
// 	}

// 	// Process the email verification requested event.
// 	return nil
// }
