// Package actor implements a self-contained actor which responds to incoming commands and events.
package actor

import (
	"context"

	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
)

// Actor represents an self-contained actor, following Hewitt's Actor Model
type Actor interface {
	// Submit publishes @c onto the command bus
	Submit(*aggregate.Command) error

	// Publish publishes @e onto the event bus
	Publish(event.Event) error

	// Action attempts to submit an @action to the internal command bus
	Action(func() error) <-chan error

	// Shutdown shuts down the actor context/loop
	Shutdown() error

	// IsActive returns true as long as the actor is able to accept commands/events
	IsActive() bool

	// Refs returns the number of active references (>= 1: active, 0: dead)
	Refs() uint32

	// Context returns the internal context. Useful to add nested/child contexts.
	Context() context.Context
}
