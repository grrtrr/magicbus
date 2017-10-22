package actor

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/eapache/channels"
	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
	"github.com/pkg/errors"
)

// GLOBAL VARIABLES
var logger = logrus.WithField("module", "actor")

// ErrShutdown is returned if the actor is no longer accepting events/commands
var ErrShutdown = errors.New("processing loop terminated")

// New instantiates a new actor in running state
// @ctx:     top-level cancellation context
// @cmdHdlr: called when a Command arrives on the Command Channel
// @evtHdlr: called when an Event arrives on the Event Channel
// @ready:   whether @cmdHdlr is ready to run immediately -  can be unblocked via ServiceReady{} event
func New(ctx context.Context, cmdHdlr func(*aggregate.Command), evtHdlr func(event.Event), ready bool) Actor {
	var a = &actor{
		refcnt:      1, // new instances always start with a reference count of 1
		actionChan:  make(chan func()),
		errChan:     make(chan error),
		eventChan:   channels.NewInfiniteChannel(),
		commandChan: channels.NewInfiniteChannel(),
	}
	a.ctx, a.cancel = context.WithCancel(ctx)

	if evtHdlr == nil {
		panic("attempt to create Actor with nil Event Handler")
	} else if cmdHdlr == nil {
		panic("attempt to create Actor with nil Command Handler")
	}
	go a.loop(cmdHdlr, evtHdlr, ready)

	return a
}

// actor is the internal implementation that provides the Actor interface
type actor struct {
	// Incoming commands
	commandChan *channels.InfiniteChannel

	// Incoming events
	eventChan *channels.InfiniteChannel

	// Single action channel to process actions addressed to the Actor itself
	actionChan chan func()

	// refcnt atomically counts the number of references to this object
	refcnt uint32

	// ctx is the cancellation context, which allows to terminate the loop
	// cancel is the cancellation function to terminate @ctx
	ctx    context.Context
	cancel context.CancelFunc

	// Any internal errors are published onto the error channel
	errChan chan error
}

// Publish publishes @e onto the Event Bus of @a.
func (a *actor) Publish(e event.Event) error {
	if !a.IsActive() {
		return ErrShutdown
	}
	a.eventChan.In() <- e
	return nil
}

// Submit submits @c onto the Command Bus of @a.
func (a *actor) Submit(c *aggregate.Command) error {
	if c == nil {
		return errors.Errorf("attempt to submit a nil command")
	} else if !a.IsActive() {
		return ErrShutdown
	}
	a.commandChan.In() <- c
	return nil
}

// Action puts @action (to change @a's internal state) onto the internal commandbus.
// Returns an error channel. Reading from channel causes synchronous processing.
func (a *actor) Action(action func() error) <-chan error {
	var errCh = make(chan error, 1)

	if !a.IsActive() {
		errCh <- ErrShutdown
	} else {
		a.actionChan <- func() { errCh <- action() }
	}
	return errCh
}

// Context exposes the internal context to allow creation of child contexts
func (a *actor) Context() context.Context {
	return a.ctx
}

// Err exposes @errChan as read-only channel to the outside
func (a *actor) Err() <-chan error {
	return a.errChan
}

// IsActive returns true if @a is still able to process events/commands
func (a *actor) IsActive() bool {
	select {
	case <-a.ctx.Done():
		return false
	default:
		return true
	}
}

// Reference counter access:
//   * each actor starts with a reference count of 1
//   * Refs() == 1 means the loop is running
//   * Refs() == 0 means the actor is dead
//   * Refs()  > 1 means this actor is referenced by other objects
func (a *actor) Refs() uint32      { return atomic.LoadUint32(&a.refcnt) }
func (a *actor) incRefcnt() uint32 { return atomic.AddUint32(&a.refcnt, 1) }
func (a *actor) decRefcnt() uint32 { return atomic.AddUint32(&a.refcnt, ^uint32(0)) }

// Shutdown shuts down the actor context/loop
func (a *actor) Shutdown() error {
	if !a.IsActive() {
		return ErrShutdown
	}

	a.cancel()
	// NB: do not wait here, since if this function is called from within a
	//     cmdHdlr, we have a deadlock situation. Rely on loop() to ensure
	//     that the reference count at end is 1, and then decremented to 0.
	return nil
}

// loop runs until a's context is canceled
func (a *actor) loop(cmdHdlr func(*aggregate.Command), evtHdlr func(event.Event), ready bool) {
	const (
		// Time to wait for other objects to release reference to actor
		refCntWaitIntvl = 1 * time.Second

		// Maximum number of @refCntWaitIntvl intervals to wait before giving up
		refCntMaxWaits = 60
	)
	var commandChan <-chan interface{}

	if ready {
		commandChan = a.commandChan.Out()
	}

	for a.IsActive() {
		select {
		case action := <-a.actionChan:
			if action != nil {
				action()
			}
		case e, ok := <-a.eventChan.Out():
			if !ok || e == nil {
				break
			} else if evt, ok := e.(event.Event); !ok {
				logger.Errorf("non-Event %v on Event channel", e)
				a.errChan <- errors.Errorf("non-Event %v on Event channel", e)
			} else {
				// The ServiceReady event serves to unblock the command channel.
				if _, ok = e.(*event.ServiceReady); ok {
					commandChan = a.commandChan.Out()
				}
				evtHdlr(evt)
			}
		case c, ok := <-commandChan:
			if !ok || c == nil {
				break
			} else if cmd, ok := c.(*aggregate.Command); !ok {
				logger.Errorf("non-Command %v on Command channel", c)
				a.errChan <- errors.Errorf("non-Command %v on Command channel", c)
			} else {
				cmdHdlr(cmd)
			}
		case <-a.ctx.Done(): // will be caught by a.IsActive()
		}
	}

	//
	// Clean-up
	//
	for cnt := 0; a.Refs() > 1; cnt++ {
		if cnt >= refCntMaxWaits {
			logger.Fatalf("timed out waiting for active references (%d) to reach 1", a.Refs())
		}
		time.Sleep(refCntWaitIntvl)
	}

	// Only close the input queues when no more events/commands can be queued
	a.eventChan.Close()
	a.commandChan.Close()

	close(a.errChan)

	// Drain output channels to terminate the internal goroutines used by InfiniteChannel
	for range a.eventChan.Out() {
	}
	for range a.commandChan.Out() {
	}

	a.decRefcnt() // Now the reference count should be 0
}
