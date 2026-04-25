package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/No2004LTC/gopher-social-ecom/internal/domain"
	segmentio "github.com/segmentio/kafka-go"
)

type interactionMQ struct {
	writer *segmentio.Writer
}

func NewInteractionMQ(brokerURL string) domain.MessageQueue {
	w := &segmentio.Writer{
		Addr:         segmentio.TCP(brokerURL),
		Topic:        "user_interactions",
		Balancer:     &segmentio.LeastBytes{},
		BatchSize:    1,
		BatchTimeout: 10 * time.Millisecond,
		Async:        true,
	}
	return &interactionMQ{writer: w}
}

// PublishInteractionEvent
func (m *interactionMQ) PublishInteractionEvent(ctx context.Context, event *domain.InteractionEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = m.writer.WriteMessages(ctx,
		segmentio.Message{
			Key:   []byte(fmt.Sprintf("%d", event.UserID)),
			Value: body,
		},
	)
	return err
}
