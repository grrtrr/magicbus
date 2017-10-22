package event

import "github.com/grrtrr/magicbus/aggregate"

// Event represents a Domain Event
type Event interface {
	// Source specifies origin of this event (may be empty)
	Source() aggregate.ID

	// Dest is intended destination aggregate (may be also be empty)
	Dest() aggregate.ID
}

// event.Handler receives an event to process.
type Handler func(Event)

// EventHandler is an optional interface implemented by Aggregates, to process events
// whose destination Dest() equals the ID of the Aggregate.
type EventHandler interface {
	HandleEvent(Event)
}
