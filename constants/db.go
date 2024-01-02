package constants

const (
	DefaultMaxConnections = 10

	ConnectionStatePending           = "pending"
	ConnectionStatePendingIncomplete = "incomplete"
	ConnectionStateReady             = "ready"
	ConnectionStateUpdating          = "updating"
	ConnectionStateDeleting          = "deleting"
	ConnectionStateDisabled          = "disabled"
	ConnectionStateError             = "error"
)

// ConnectionStates is a handy array of all states
var ConnectionStates = []string{
	ConnectionStatePending,
	ConnectionStateReady,
	ConnectionStateUpdating,
	ConnectionStateDeleting,
	ConnectionStateError,
}
