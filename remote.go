package magicbus

import (
	"context"

	"github.com/grrtrr/magicbus/aggregate"
	"github.com/grrtrr/magicbus/event"
	"github.com/pkg/errors"
)

/*
 * Dealing with remote commands and events.
 */
// remoteSubmit sends @cmd to the remote bus specified by cmd.AggregateID()
func remoteSubmit(ctx context.Context, cmd *aggregate.Command) error {
	return errors.Errorf("remoteSubmit NOT IMPLEMENTED YET: integrete your remote method call here")
}

// remotePublish forwards @evt to the remote event bus specified by @evt.To
func remotePublish(ctx context.Context, evt event.Event) error {
	return errors.Errorf("remotePublish NOT IMPLEMENTED YET: integrete your remote method call here")
}
