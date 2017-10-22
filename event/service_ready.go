package event

import (
	"fmt"

	"github.com/grrtrr/magicbus/aggregate"
)

// ServiceReady is sent to enable command processing in Aggregates.
//
// This applies only to Aggregates which were initially registed with
// a 'ready=false' flag, meaning that incoming commands are queued, but
// will not be processed until a ServiceReady event tells the Aggregate
// that it is now time to do so.
type ServiceReady struct {
	Aggregate aggregate.ID // Aggregate to unblock
}

func (s *ServiceReady) Source() aggregate.ID { return s.Aggregate }
func (s *ServiceReady) Dest() aggregate.ID   { return s.Source() }

func (s ServiceReady) String() string {
	return fmt.Sprintf("ServiceReady(%s)", s.Aggregate.ID)
}
