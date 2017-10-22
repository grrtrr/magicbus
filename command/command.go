// Package command defines the base structure of a command.
package command

import (
	"context"
	"fmt"
)

// Command carries a struct or string in Data() and is associated with a Job().
type Command interface {
	// Wraps the specific kind of command (named struct or non-empty string)
	Data() interface{}

	// Cancellation context for this command
	Context() context.Context
}

// command.Result groups the result of a command, and the potential error on failure.
type Result struct {
	Result interface{} // Return value(s), may be nil
	Err    error       // Non-nil value if the command failed
}

func (r Result) String() string {
	if r.Err != nil {
		return fmt.Sprintf("err = %s", r.Err)
	} else if r.Result == nil {
		return "OK"
	} else if s, ok := r.Result.(string); ok {
		if s == "" {
			return "OK"
		}
		return s
	}
	return fmt.Sprintf("result = %v", r.Result)
}
