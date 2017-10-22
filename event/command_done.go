package event

import (
	"errors"
	"fmt"

	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/command"
)

// CommandDone is the '<cmd>Done' event published whenever a command completes.
type CommandDone struct {
	Src aggregate.ID // Aggregate reporting this event, the status of a command just run
	Dst aggregate.ID // Intended destination Aggregate (issuer of the command to be run)

	Desc   string      // Descriptive text (used for logging)
	Data   interface{} // The command data embedded in the original command
	Status string      // Result: success status as string
	Error  string      // Result: stringified error (empty means no error)
}

// NewCmdDone is a convenience wrapper that fills in an event from @a and @cmd
func NewCmdDone(src aggregate.ID, cmd *aggregate.Command, result interface{}, err error) Event {
	var cd = &CommandDone{
		Src:  src,
		Dst:  cmd.Source(),
		Data: cmd.Data(),
		Desc: cmd.Type(),
	}

	if result != nil {
		cd.Status = fmt.Sprint(result) // FIXME: stringification
	}

	if err != nil {
		cd.Error = err.Error()
	}
	return cd
}

func (c *CommandDone) Source() aggregate.ID { return c.Src }
func (c *CommandDone) Dest() aggregate.ID   { return c.Dst }
func (c *CommandDone) Result() command.Result {
	if c.Error != "" {
		return command.Result{Result: c.Status, Err: errors.New(c.Error)}
	}
	return command.Result{Result: c.Status}
}

func (c CommandDone) String() string {
	return fmt.Sprintf("CommandDone(%s, %s)", c.Desc, c.Result())
}
