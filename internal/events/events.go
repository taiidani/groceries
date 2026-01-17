// Package events provides Server-Sent Events (SSE) functionality for real-time updates.
// It implements event formatting and streaming according to the SSE specification.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
)

type PubSub interface {
	Subscribe(ctx context.Context, channels ...string) <-chan Event
	Publish(ctx context.Context, channel string, data fmt.Stringer) error
}

type EventServer struct {
	client *redis.Client
}

func NewRedisPubSub(client *redis.Client) PubSub {
	return &EventServer{
		client: client,
	}
}

func (srv *EventServer) Subscribe(ctx context.Context, channels ...string) <-chan Event {
	sub := srv.client.Subscribe(ctx, channels...)

	ret := make(chan Event)

	go func() {
		receiver := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-receiver:
				evt := Event{}
				if err := json.Unmarshal([]byte(msg.Payload), &evt); err != nil {
					slog.Warn("Unmarshal of event failure", "error", err)
				}

				ret <- evt
			}
		}
	}()

	return ret
}

func (srv *EventServer) Publish(ctx context.Context, channel string, data fmt.Stringer) error {
	evt := Event{
		Event: channel,
		Data:  data,
	}

	payload, err := json.Marshal(evt)
	if err != nil {
		return fmt.Errorf("event marshaling error: %w", err)
	}

	cmd := srv.client.Publish(ctx, channel, payload)
	return cmd.Err()
}
