// Package query implements the read-side of our EventSourcing system,
// providing structure and functions to retrieve up-to-date information
// from an Aggregate.
package query

import "github.com/grrtrr/magicbus/aggregate"

// query.Type indicates which kind of information (attributes) we are interested in
type Type string

// query.Argument specifies the input to a Query
type Argument interface {
	// QueryType indicates the action of the query - what it is asking for
	QueryType() Type

	// Aggregate specifies the target (which aggregate) to query
	AggregateID() aggregate.ID
}

// Handler provides Query @results based on @args
type Handler interface {
	Query(args Argument) (results interface{}, err error)
}
