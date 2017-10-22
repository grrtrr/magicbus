package aggregate

import (
	"testing"
)

func simpleNewCommand(a ID, cmdData interface{}) (*Command, error) {
	return NewCommand(a, a, cmdData)
}

func TestCommandCreation(t *testing.T) {
	testAggregateID := ID{Type: 42}

	// Must not register a nil command
	if _, err := simpleNewCommand(testAggregateID, nil); err == nil {
		t.Fatalf("expected error attempting to register a nil command, but got no error")
	}

	// Must not register a non-struct command
	if _, err := simpleNewCommand(testAggregateID, int(42)); err == nil {
		t.Fatalf("expected error attempting to register a non-struct/non-string type, but got no error")
	}

	// Must not attempt to register an empty string as a command
	if _, err := simpleNewCommand(testAggregateID, ""); err == nil {
		t.Fatalf("expected error attempting to register an empty string, but got no error")
	}

	// Must not register an anonymous command
	if _, err := simpleNewCommand(testAggregateID, struct{}{}); err == nil {
		t.Fatalf("expected error attempting to register an anonymous command, but got no error")
	}

	// Another anonymous command
	var x struct{ foo, bar string }
	if _, err := simpleNewCommand(testAggregateID, x); err == nil {
		t.Fatalf("expected error attempting to register an anonymous command, but got no error")
	}
	if _, err := simpleNewCommand(testAggregateID, &x); err == nil {
		t.Fatalf("expected error attempting to register an anonymous command, but got no error")
	}

	// Named structs
	y := namedCommand1{}
	if c, err := simpleNewCommand(testAggregateID, y); err != nil {
		t.Fatalf("failed to create namedCommand1: %s", err)
	} else if n := c.Type(); n != "namedCommand1" {
		t.Fatalf("error creating namedCommand1: name %q does not match", n)
	}

	// Same using a pointer
	if c, err := simpleNewCommand(testAggregateID, &y); err != nil {
		t.Fatalf("failed to create namedCommand1: %s", err)
	} else if n := c.Type(); n != "namedCommand1" {
		t.Fatalf("error creating namedCommand1: name %q does not match", n)
	}

	z := namedCommand2{}
	if c, err := simpleNewCommand(testAggregateID, &z); err != nil {
		t.Fatalf("failed to create namedCommand2: %s", err)
	} else if n := c.Type(); n != "namedCommand2" {
		t.Fatalf("error creating namedCommand2: name %q does not match", n)
	}
}

type namedCommand1 struct {
	field1 string
	field2 string
	field3 int
}

type namedCommand2 struct {
	foo, bar []byte
	field3   int
}
