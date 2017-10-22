package aggregate

// Aggregate represents an aggregate entity (what we call subsystem in Syntropy).
type Aggregate interface {
	// Returns the cluster-unique ID of this Aggregate
	AggregateID() ID

	// HandleCommand lets the Aggregate handle @Command
	// @next:   if not nil, returns next command-in-sequence to complete
	// @result: (only if @next=nil) returns result of operation
	// @err:    error value (@next/@result are ignored in this case)
	HandleCommand(*Command) (next *Command, result interface{}, err error)
}
