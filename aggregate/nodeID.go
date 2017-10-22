package aggregate

/*
 * Node-unique ID
 */

// GLOBAL VARIABLES
var (
	// ID of this node. This typically is the primary IP address, but can also be e.g. /etc/machine-id
	nodeID string = "UNKNOWN"
)

// Allow other packages to retrieve the ID of this node
func NodeID() string {
	return nodeID
}

// Set node ID to @id.
func SetNodeID(id string) {
	nodeID = id
}
