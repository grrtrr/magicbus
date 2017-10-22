package magicbus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
	"github.com/pkg/errors"
)

// GLOBALS
var (
	cmdh       = simpleCommand{"hello command"}
	cmd1       = simpleCommand{"command one"}
	cmd2       = simpleCommand{"command two"}
	cmd1Failed = errors.New("command 1 failed")
)

func init() {
	Init(context.Background())
	aggregate.SetNodeID("testNode")
}

func TestNewMagicBus(t *testing.T) {
	var m = NewMagicBus(context.Background())

	if r := m.Refs(); r != 1 {
		t.Fatalf("reference count (%d) not 1 after start", r)
	} else if !m.IsActive() {
		t.Fatalf("bus is inactive after start: %s", m)
	}

	// Attempt to register aggregate with empty aggregate ID should fail
	a := &testAggregate{t: t}
	if err := m.Register(a, true); err == nil {
		t.Fatalf("expected an error when registering with empty aggregate ID, but got nil")
	}

	// Attempt to register an empty command list should fail
	if err := m.Register(a, true); err == nil {
		t.Fatalf("expected error when registering with an empty command list, but got nil")
	}

	a.id = aggregate.NewID(aggregate.ResourceType_CPU, "amd64")
	if err := m.Register(a, true); err != nil {
		t.Fatalf("failed to register new aggregate: %s", err)
	}
	t.Logf("new magic bus %s", m)

	// Test command submission on local bus
	RegisterAggregate(a, true) // register with the local bus instance

	command1, err := aggregate.NewCommand(a.AggregateID(), a.AggregateID(), cmd1)
	if err != nil {
		t.Fatalf("failed to set up cmd1: %s", err)
	} else if err := m.Submit(command1); err != nil {
		t.Fatalf("failed to submit command %s: %s", command1, err)
	}

	// Launch test
	res := Launch(context.TODO(), command1)
	if res.Err != nil {
		t.Fatalf("launch failed: %s", res.Err)
	}
	t.Logf("result of launching %s: %s", cmd1, res)

	//
	// Event subscription tests
	//
	te := mkTestEvent(a.AggregateID(), a.AggregateID(), "Test event")

	// 1. One-off subscription: multiple events result in only 1 handler call
	teHdlr := func(e event.Event) {
		t.Logf("first event handler received %s", e)
	}
	id, err := m.observer(teHdlr)
	if err != nil {
		t.Fatalf("failed to subscribe %s: %s", te, err)
	}
	t.Logf("%s new event subscription %s", a.AggregateID(), id)

	// 2. First subscription
	teHdlr2 := func(e event.Event) {
		t.Logf("second event handler received %s", e)
	}
	id1, err := m.observer(teHdlr2)
	if err != nil {
		t.Fatalf("failed to subscribe %s: %s", te, err)
	} else if id1 == id {
		t.Fatalf("event subscriptions not unique: %s", id)
	}
	t.Logf("%s new event subscription %s", a.AggregateID(), id)

	// Publish event, and wait some time for the handlers to respond
	if err := m.Publish(te); err != nil {
		t.Fatalf("failed to publish test event: %s", err)
	}

	// Publish again
	m.Publish(te)

	// Wait for handlers to settle and event to reach aggregate before unsubscription
	time.Sleep(1 * time.Second)

	if err := m.unsubscribe(id); err != nil {
		t.Fatalf("failed to unsubscribe first event handler: %s", err)
	} else if err := m.unsubscribe(id1); err != nil {
		t.Fatalf("failed to unsubscribe second event handler: %s", err)
	}

	// This should now report 1 aggregate and 0 subscriptions:
	t.Logf("magic bus now: %s", m)

	// Shutdown test, with a minimal wait time for Shutdown to kick in.
	m.Shutdown()
	time.Sleep(100 * time.Millisecond)
	if m.IsActive() {
		t.Fatalf("bus is active after shutdown")
	} else if r := m.Refs(); r != 0 {
		t.Fatalf("reference count (%d) not 0 after shutdown", r)
	}
}

// Test command
func mkTestCommand(id aggregate.ID, typ string) *aggregate.Command {
	c, err := aggregate.NewCommand(id, id, typ)
	if err != nil {
		panic(fmt.Sprintf("failed to create command: %s", err))
	}
	return c
}

// We only allow structs as Command type, since the name of the struct
// doubles as command name (type).
type simpleCommand struct {
	name string
}

// Test event
func mkTestEvent(src, dst aggregate.ID, kind string) event.Event {
	return &testEvent{src: src, dst: dst, kind: kind}
}

type testEvent struct {
	src, dst aggregate.ID
	kind     string
}

func (t *testEvent) Source() aggregate.ID { return t.src }
func (t *testEvent) Dest() aggregate.ID   { return t.dst }
func (t testEvent) String() string        { return fmt.Sprintf("%s => %s: %q", t.src, t.dst, t.kind) }

// Test aggregate
type testAggregate struct {
	id aggregate.ID
	t  *testing.T
}

func (t *testAggregate) AggregateID() aggregate.ID {
	return t.id
}

// FIXME: make an aggregate that fails upon HandleCommand()
func (t *testAggregate) HandleCommand(cmd *aggregate.Command) (*aggregate.Command, interface{}, error) {

	t.t.Logf("%s handling %s command", t.id, cmd)
	// NB: CommandDone will be published by aggregate_actor anyway
	return nil, nil, nil
}
