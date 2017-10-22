package magicbus

import (
	"context"
	"fmt"

	"github.com/grrtrr/magicbus/actor"
	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
	"github.com/pkg/errors"
)

// MagicBus serializes event/command notification on behalf of aggregates, allowing
// event observers to subscribe to asynchronous/immediate event notifications.
type MagicBus struct {
	// Internal Actor object
	actor.Actor

	// List of command-handling aggregates (map { AggregateID -> aggregateActor })
	aggregates map[aggregate.ID]*aggregateActor

	// List of event observers (map { SubscriptionID -> event.Handler })
	observers map[string]event.Handler
}

// NewMagicBus instantiates a new bus instance ready to process commands/events.
func NewMagicBus(ctx context.Context) *MagicBus {
	m := &MagicBus{
		aggregates: map[aggregate.ID]*aggregateActor{},
		observers:  map[string]event.Handler{},
	}
	m.Actor = actor.New(ctx, m.commandHandler, m.eventHandler, true)
	return m
}

// command-processing callback
func (m *MagicBus) commandHandler(cmd *aggregate.Command) {
	// Try most-specific match (Type + Node + ID) first
	if ag, ok := m.aggregates[cmd.Dest()]; ok {
		if err := ag.Submit(cmd); err != nil {
			logger.Errorf("%s: failed to submit %v: %s", ag.AggregateID(), cmd, err)
		}
		return
	} else if cmd.Dest().ID != "" {
		// If there is no specific instance, try the general subsystem (ID == "")
		if ag, ok := m.aggregates[aggregate.NewID(cmd.Dest().Type, "")]; ok {
			if err := ag.Submit(cmd); err != nil {
				logger.Errorf("%s: failed to submit %v: %s", ag.AggregateID(), cmd, err)
			}
			return
		}
	}

	// No match means we are unable to handle a legitimate command.
	logger.Fatalf("magicbus: no aggregate handler was interested in %s", cmd)
}

// eventHandler is called my m.actor for each incoming event
func (m *MagicBus) eventHandler(e event.Event) {
	// 1. Aggregates receive all events directed to them.
	if ag, ok := m.aggregates[e.Dest()]; ok {
		if err := ag.Publish(e); err != nil {
			logger.Warningf("%s: failed to publish %v: %s", ag.AggregateID(), e, err)
		}
	}

	// 2. Observers are handled in parallel.
	for _, handler := range m.observers {
		eventHandler := handler // avoid loop variable alias
		go eventHandler(e)
	}
}

func (m *MagicBus) String() string {
	var resChan = make(chan string, 1)

	if err := <-m.Action(func() error {
		resChan <- fmt.Sprintf("bus (aggregates: %d, subscriptions: %d)", len(m.aggregates), len(m.observers))
		return nil
	}); err != nil {
		return fmt.Sprintf("bus in error: %s", err)
	}
	return <-resChan
}

// Register registers @a to handle commands on the local bus.
// @ready: whether the aggregate is ready to process commands right away
func (m *MagicBus) Register(a aggregate.Aggregate, ready bool) error {
	if a == nil {
		return errors.Errorf("attempt to register a nil Aggregate")
	} else if a.AggregateID().IsZero() {
		return errors.Errorf("attempt to register an Aggregate with an empty AggregateID")
	}

	return <-m.Action(func() error {
		logger.Debugf("magicbus: registering %s", a.AggregateID())

		// Allow duplicate registration for robustness, reusing the first one.
		if _, exists := m.aggregates[a.AggregateID()]; !exists {
			m.aggregates[a.AggregateID()] = newAggregateActor(m.Context(), a, ready)
		}
		return nil
	})
}

// Unregister removes @a from the bus
func (m *MagicBus) Unregister(id aggregate.ID) error {
	return <-m.Action(func() error {
		logger.Debugf("magicbus: de-registering %s", id)

		ag, ok := m.aggregates[id]
		if !ok {
			return nil
		}
		delete(m.aggregates, id)
		return ag.Shutdown()
	})
}
