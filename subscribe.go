package magicbus

import (
	"github.com/grrtrr/magicbus/event"
	uuid "github.com/satori/go.uuid"
)

// Each subscription is identified by a cluster-unique ID
type SubscriptionID uuid.UUID

// NewSubscriptionID returns a new, cluster-unique subscription ID
func NewSubscriptionID() SubscriptionID {
	return SubscriptionID(uuid.NewV1())
}

func (s SubscriptionID) String() string {
	return uuid.UUID(s).String()
}

func (s SubscriptionID) IsZero() bool {
	return uuid.Equal(uuid.UUID(s), uuid.Nil)
}

// Observer subscribes @hdlr to receive immediate notification of events.
func Observer(hdlr event.Handler) (SubscriptionID, error) {
	return localBus.observer(hdlr)
}

// Unsubscribe removes subscription @id from the local bus.
func Unsubscribe(id SubscriptionID) error {
	return localBus.unsubscribe(id)
}

// Add new observer to @m
func (m *MagicBus) observer(hdlr event.Handler) (SubscriptionID, error) {
	var id = NewSubscriptionID()

	return id, <-m.Action(func() error {
		m.observers[id.String()] = hdlr
		return nil
	})
}

// Remove subscription records of @id
func (m *MagicBus) unsubscribe(id SubscriptionID) error {
	return <-m.Action(func() error {
		delete(m.observers, id.String())
		return nil
	})
}
