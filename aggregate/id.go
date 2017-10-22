package aggregate

import (
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"
)

// ID is used to identify a subsystem uniquely across the entire cluster
type ID struct {
	// Node on which this Aggregate resides
	Node string /* FIXME: NodeID */

	// Subsystem that this aggregate belongs to (Aggregate Root)
	Type ResourceType

	// Unique ID of this aggregate on this node
	// This may be empty if the entity is the aggregate root.
	ID string
}

func (a ID) String() string {
	if a.Node == "" {
		return a.Resource()
	}
	return fmt.Sprintf("%s.%s", a.Node, a.Resource())
}

// Resource specifies the resource(s) (subsystem + ID) identified by @a.
func (a ID) Resource() string {
	if a.ID == "" {
		return a.Type.String()
	}
	return fmt.Sprintf("%s.%s", a.Type, a.ID)
}

// IsZero returns true if @a is not sufficiently specified (initialized).
func (a ID) IsZero() bool {
	return a.Node == "" || a.Type == ResourceType_INVALID_RESOURCE
}

// IsLocal returns true if @a points to this/the local node.
func (a ID) IsLocal() bool {
	return a.Node == NodeID()
}

// NewID returns a new aggregate.ID for this node.
// @aggregateRoot: the subsystem (name of bounded context) of this aggregate
// @entityID:      within the @aggregateRoot, unique ID of the aggregate on this node
func NewID(aggregateRoot ResourceType, entityID string) ID {
	return ID{Node: NodeID(), Type: aggregateRoot, ID: entityID}
}

// Implements encoding.TextMarshaler
func (i ID) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// Implements encoding.TextUnmarshaler
func (i *ID) UnmarshalText(data []byte) (err error) {
	*i = ID{} // zero out fields in case @i was reused

	fields := strings.Split(string(data), ".")
	switch len(fields) {
	case 6: // node is an IP address
		i.Node = strings.Join(fields[:4], ".")
		if net.ParseIP(i.Node) == nil {
			return errors.Errorf("invalid node IP %q in %s", i.Node, string(data))
		}
		i.Type, err = ResourceTypeFromString(fields[4])
		i.ID = fields[5]
	case 5: // node is an IP address
		i.Node = strings.Join(fields[:4], ".")
		if net.ParseIP(i.Node) == nil {
			return errors.Errorf("invalid node IP %q in %s", i.Node, string(data))
		}
		i.Type, err = ResourceTypeFromString(fields[4])
	case 4: // solitary IP address
		i.Node = string(data)
		if net.ParseIP(i.Node) == nil {
			return errors.Errorf("invalid node IP %q in %s", i.Node, string(data))
		}
	case 3: // node is a name without dot in it
		i.ID = fields[2]
		fallthrough
	case 2: // Node.Resource
		i.Node = fields[0]
		i.Type, err = ResourceTypeFromString(fields[1])
	case 1: // Resource only
		i.Type, err = ResourceTypeFromString(fields[0])
	default:
		return errors.Errorf("invalid Aggregate ID %q", string(data))
	}
	return err
}
