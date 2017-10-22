package magicbus

import (
	"context"

	"github.com/grrtrr/magicbus/actor"
	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
)

// aggregateActor serializes command/event handling on behalf of a registered Aggregate
type aggregateActor struct {
	// Aggregate handled by this actor
	aggregate.Aggregate

	// Internal actor object
	actor.Actor
}

// newAggregateActor returns an initialized new Actor
// @ctx:    Cancellation context
// @agg:    Aggregate represented by this aggregateActor
// @ready:  Whether @agg is ready to run its HandleCommand() function.
//          If set to false, can be enabled later by sending a ServiceReady event.
func newAggregateActor(ctx context.Context, agg aggregate.Aggregate, ready bool) *aggregateActor {
	a := &aggregateActor{Aggregate: agg}

	a.Actor = actor.New(ctx, a.commandHandler, a.eventHandler, ready)
	return a
}

// command-processing callback
func (a *aggregateActor) commandHandler(cmd *aggregate.Command) {
	var agId = a.AggregateID()

	// The Dest of a command identifies the matching aggregate, with the only exception
	// that a specific command (ID != "") is sent to the "general manager" (ID == "").
	if cmd.Dest() != agId && (agId.ID != "" || cmd.Dest().Type != agId.Type || cmd.Dest().Node != agId.Node) {
		logger.Errorf("%s: refusing to handle command - mismatching aggregate ID %s", a.AggregateID(), cmd.Dest())
		return
	} else if err := cmd.Context().Err(); err != nil {
		logger.Warningf("%s: command canceled (%s)", a.AggregateID(), err)
		return
	}
	nextStep, result, err := a.Aggregate.HandleCommand(cmd)

	// Emit the CommandDone event to notify the (remote) site of completion.
	// Note: agId and cmd.Dest() may differ in the case where a new, specific Aggregate is created.
	//       In this case, agID.ID=="" and cmdID.ID != "", and we have created a new Aggregate to
	//       handle the event. Thus, the _actual_ source Aggregate is cmd.Dest().
	//       If ever changing the creation of specific managers, this MUST also be updated.
	Publish(event.NewCmdDone(cmd.Dest() /* see comment above */, cmd, result, err))

	// Submit the nextStep command only _after_ publishing the events (otherwise the timing is off).
	if nextStep != nil {
		Submit(cmd.Context(), nextStep)
	}
}

// eventHandler is called by a.actor for each incoming event e whose Dest() matches the AggregateID of @a.
func (a *aggregateActor) eventHandler(e event.Event) {
	if _, ok := e.(*event.ServiceReady); ok { // ServiceReady events are not passed on any further.
		logger.Debugf("%s: ready to process commands", a.AggregateID())
	} else if eh, ok := a.Aggregate.(event.EventHandler); ok {
		eh.HandleEvent(e)
	}
}
