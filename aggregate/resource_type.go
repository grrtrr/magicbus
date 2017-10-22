package aggregate

import (
	"fmt"

	"github.com/pkg/errors"
)

// ResourceType specifies the type of an Aggregate.
type ResourceType int32

const (
	// Indicates that no valid resource type has been assigned yet
	ResourceType_INVALID_RESOURCE ResourceType = 0

	// The following are for example only
	ResourceType_CPU    ResourceType = 1
	ResourceType_MEMORY ResourceType = 2
)

// FIXME: the ResourceType typically is an enum, as sketched above with the
//        begin of the enum literals. You would define your own stringer and
//        from/to JSON methods for this type, depending on your needs.
func (r ResourceType) String() string {
	switch r {
	case ResourceType_INVALID_RESOURCE:
		return "INVALID_RESOURCE"
	case ResourceType_CPU:
		return "CPU"
	case ResourceType_MEMORY:
		return "MEMORY"
	}
	return fmt.Sprint(int32(r))
}

// ResourceTypeFromString is again project specific. It is the inverse of String().
func ResourceTypeFromString(s string) (ResourceType, error) {
	switch s {
	case "INVALID_RESOURCE":
		return ResourceType_INVALID_RESOURCE, nil
	case "CPU":
		return ResourceType_CPU, nil
	case "MEMORY":
		return ResourceType_MEMORY, nil
	}
	return ResourceType_INVALID_RESOURCE, errors.Errorf("invalid ResourceType %q", s)
}
