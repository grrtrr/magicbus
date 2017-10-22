// Package repository handles read access to query Aggregates registered with the MagicBus
package repository

import (
	"github.com/Sirupsen/logrus"
	"github.com/grrtrr/magicbus"
	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
	"github.com/grrtrr/magicbus/query"
	"github.com/pkg/errors"
)

// Global Variables
var (
	logger = logrus.WithField("module", "repository")
	// registeredAggregates is populated at package initialization time
	registeredAggregates = map[aggregate.ResourceType]query.Handler{}
)

// A repository handles information pertaining to the AggregateType() it advertises.
// Any queries relating to the advertised AggregateType() are routed to its query.Handler.
type Repository interface {
	// AggregateType returns the subsystem that this repository is reponsible for
	AggregateType() aggregate.ResourceType

	// Update is an event.Handler, called each time an event for AggregateType() arrives
	Update(event.Event)

	// All queries pertaining to AggregateType() are delegated to Handler
	query.Handler
}

// HandleQuery delegates @q to the appropriate Aggregate (query router)
func HandleQuery(q query.Argument) (results interface{}, err error) {
	if repo, ok := registeredAggregates[q.AggregateID().Type]; ok {
		return repo.Query(q)
	}
	return nil, errors.Errorf("no handler registered for %s query", q.AggregateID())
}

// RegisterQueryHandler registers @h as handling queries pertaining to @subsystem (at package initialization time)
func RegisterQueryHandler(r Repository) {
	if _, err := magicbus.Observer(func(e event.Event) {
		if e.Source().Node == aggregate.NodeID() && e.Source().Type == r.AggregateType() {
			r.Update(e)
		}
	}); err != nil {
		logger.Fatalf("could not subscribe repository for %s domain-specific events: %s", r.AggregateType(), err)
	}
	registeredAggregates[r.AggregateType()] = r

}
