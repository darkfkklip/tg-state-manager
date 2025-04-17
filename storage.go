package tgstatemanager

// UserState holds the current state name and data.
type UserState[S any] struct {
	CurrentState string
	Data         S
	PromptSent   bool // Tracks if prompt has been sent for the current state
}

// StateStorage defines the interface for storing user states.
type StateStorage[S any] interface {
	Get(id int64) (UserState[S], bool, error)
	Set(id int64, state UserState[S]) error
}