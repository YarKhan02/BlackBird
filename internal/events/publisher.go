package events

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, payload any) error
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, topic string, payload any) error {
	return nil
}
