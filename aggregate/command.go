package aggregate

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// Command is an implementation of command.Command
type Command struct {
	// dst designates the receiver of this command
	dst ID

	// src is the subystem issuing this command
	src ID

	// Args contains the command-specific struct or string
	args interface{}

	// Cancellation context
	ctx context.Context
}

// WithContext adds @ctx to @c and returns the transformed result and cancel function.
func (c *Command) WithContext(ctx context.Context) (c1 *Command, cancel context.CancelFunc) {
	c.ctx, cancel = context.WithCancel(ctx)
	return c, cancel
}

// NewLocalCommand is the simplest use case: local aggregate, no job tracking.
func NewLocalCommand(aggregate ID, cmdData interface{}) (*Command, error) {
	return NewCommand(aggregate, aggregate, cmdData)
}

// NewCommand creates a command with job tracking.
// @src, dst: source/destination of the command
// @cmdData:  either
//            a) a (pointer to a) struct - in this case Type() is the type-name of the struct, or
//            b) a non-empty string - in this case Type() is the content of the string.
func NewCommand(src, dst ID, cmdData interface{}) (*Command, error) {
	var t = getType(cmdData)

	if t == nil {
		return nil, errors.Errorf("attempt to submit a nil command")
	}

	switch t.Kind() {
	case reflect.Struct:
		if t.Name() == "" {
			return nil, errors.Errorf("attempt to submit anonymous struct as command")
		}
	case reflect.String:
		if fmt.Sprint(cmdData) == "" {
			return nil, errors.Errorf("attempt to submit an empty string as command")
		}
	default:
		return nil, errors.Errorf("attempt to submit %s %#+v as command", t.Kind(), cmdData)
	}

	return &Command{src: src, dst: dst, args: cmdData, ctx: context.Background()}, nil

}

// IsLocal returns true if @cmd is not meant to leave the local bus.
func (c *Command) IsLocal() bool  { return c.src.IsLocal() && c.dst.IsLocal() }
func (c *Command) String() string { return c.Type() }

// ToJSON represents @c as a JSON string
func (c *Command) ToJSON() string {
	if b, err := json.Marshal(c.Data()); err != nil {
		return fmt.Sprintf("ERROR: failed to serialize %s: %s", c.Type(), err)
	} else {
		return fmt.Sprintf("%s %s", c.Type(), string(b))
	}
}

// Getters (no Setters)
func (c *Command) Source() ID               { return c.src }
func (c *Command) Dest() ID                 { return c.dst }
func (c *Command) Context() context.Context { return c.ctx }

// Implements command.Command
func (c *Command) Data() interface{} { return c.args }

// Type returns a description of @cmd
func (c *Command) Type() string {
	switch t := getType(c.Data()); t.Kind() {
	case reflect.String:
		return c.Data().(string)
	case reflect.Struct:
		return t.Name()
	default:
		return fmt.Sprintf("invalid command %T %#+v", c.Data(), c.Data())
	}
}

// getType is a helper that extracts the underlying type of @cmd
func getType(cmd interface{}) reflect.Type {
	var t = reflect.TypeOf(cmd)

	if t != nil {
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}
	return t
}
