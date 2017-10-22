// Package magicbus implements a combined EventBus/CommandBus queueing service which guarantees that event/command handlers
// of registered aggregates are run in serialized (non-current) order.
package magicbus

import (
	"context"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/grrtrr/magicbus/actor"
	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/command"
	"github.com/grrtrr/magicbus/event"
	"github.com/pkg/errors"
)

// GLOBAL VARIABLES
var (
	localBus *MagicBus
	logger   = logrus.WithField("module", "magicbus")
)

// Init allocates the %localBus.
func Init(ctx context.Context) {
	localBus = NewMagicBus(ctx)
}

// Launch takes command @data, turns it into a Command, and submits it to the local bus.
// The result of the command (via the CommandDone event) is reported via the error channel.
func Launch(ctx context.Context, cmd *aggregate.Command) command.Result {
	var resultCh = make(chan command.Result, 1)

	id, err := Observer( // Perform a one-off subscription for the CommandDone event.
		func(e event.Event) {
			if cd, ok := e.(*event.CommandDone); ok && e.Dest() == cmd.Source() {
				resultCh <- cd.Result()
			}
		},
	)
	if err != nil {
		return command.Result{Err: errors.Errorf("failed to subscribe to %s CommandDone event: %s", cmd.Type(), err)}
	}
	defer Unsubscribe(id)

	if err = Submit(ctx, cmd); err != nil {
		return command.Result{Err: errors.Errorf("failed to submit %s: %s", cmd.Type(), err)}
	}

	select {
	case ret := <-resultCh:
		return ret
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return command.Result{Err: errors.Errorf("timed out waiting for %s to complete", cmd.Type())}
		}
		return command.Result{Err: ctx.Err()}
	case <-cmd.Context().Done():
		return command.Result{Err: errors.Errorf("command %s canceled: %s", cmd.Type(), cmd.Context().Err())}
	}
}

// LaunchWait is a variation of Launch which takes a timeout @maxWait instead of a context.
func LaunchWait(cmd *aggregate.Command, maxWait time.Duration) command.Result {
	ctx, _ := context.WithTimeout(cmd.Context(), maxWait)
	return Launch(ctx, cmd)
}

// Submit @cmd to the local bus or forward it to a remote bus.
func Submit(ctx context.Context, cmd *aggregate.Command) error {
	if !cmd.Dest().IsLocal() {
		return remoteSubmit(ctx, cmd)
	}
	return localBus.Submit(cmd)
}

// Publish @evt on the local bus, or pass it on to shcomm as STATUS message.
func Publish(evt event.Event) {
	if err := func() error {
		if !evt.Dest().IsZero() && !evt.Dest().IsLocal() {
			return remotePublish(localBus.Context(), evt)
		}
		return localBus.Publish(evt)
	}(); err != nil && err != actor.ErrShutdown {
		logger.Errorf("%s: failed to publish event (%s): %s", evt, err)
	}
}

// RegisterAggregate registers @a to handle commands on the local bus.
// @ready: whether the aggregate is ready to process commands right away
func RegisterAggregate(a aggregate.Aggregate, ready bool) {
	if err := localBus.Register(a, ready); err != nil {
		logger.Fatalf("%s: registration failed: %s", a.AggregateID(), err)
	}
}

// UnregisterAggregate removes Aggregate @id from the bus.
func UnregisterAggregate(id aggregate.ID) {
	if err := localBus.Unregister(id); err != nil {
		logger.Fatalf("%s: de-registration failed: %s", id, err)
	}
}
