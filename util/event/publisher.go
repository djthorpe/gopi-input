package event

import (
	"github.com/djthorpe/gopi"
)

type Publisher struct {
}

// Return a channel for subscriptions
func (this *Publisher) Subscribe() <-chan gopi.Event {
	return nil
}

// Return a channel for unsubscribing
func (this *Publisher) Unsubscribe(channel <-chan gopi.Event) {
	// Do nothing
}
