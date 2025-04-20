package tgstatemanager

import (
	"errors"
	"fmt"
)

var (
	// ErrValidation indicates a validation failure, keeping the current state.
	ErrValidation = errors.New("validation error")
	// ErrDuplicateState is returned when attempting to add a state with a name that already exists.
	ErrDuplicateState = errors.New("duplicate state name")
	// ErrEmptyStateName is returned when attempting to add a state with an empty name.
	ErrEmptyStateName = errors.New("empty state name")
)

// NopState is a special state name indicating no state transition should occur.
const NopState = "<nop>"

// State defines a state in the state machine.
type State[S, U any] struct {
	Name   string
	Prompt func(update U, state *S) error           // Optional: Runs when entering the state
	Handle func(update U, state *S) (string, error) // Handles updates, returns next state
}

// StateManager manages states for Telegram bots.
type StateManager[S, U any] struct {
	states       map[string]*State[S, U]
	storage      StateStorage[S]
	keyFunc      func(update U) int64
	initialState string
}

// NewStateManager creates a new StateManager.
func NewStateManager[S, U any](storage StateStorage[S], keyFunc func(update U) int64) *StateManager[S, U] {
	return &StateManager[S, U]{
		states:  make(map[string]*State[S, U]),
		storage: storage,
		keyFunc: keyFunc,
	}
}

// Add adds states to the manager and returns an error if any duplicate state names are found.
// It returns an error if any duplicate state names are found or if any state names are empty.
func (m *StateManager[S, U]) Add(states ...*State[S, U]) error {
	for _, state := range states {
		if state.Name == "" {
			return ErrEmptyStateName
		}
		if _, exists := m.states[state.Name]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateState, state.Name)
		}
		m.states[state.Name] = state
	}

	return nil
}

// SetInitialState sets the initial state for new users.
func (m *StateManager[S, U]) SetInitialState(name string) {
	m.initialState = name
}

// Handle processes an update, managing state transitions.
func (m *StateManager[S, U]) Handle(update U) (bool, error) {
	key := m.keyFunc(update)
	userState, exists, err := m.storage.Get(key)
	if err != nil {
		return false, err
	}

	if !exists {
		userState.CurrentState = m.initialState
	}

	state, ok := m.states[userState.CurrentState]
	if !ok {
		return false, nil // Invalid state, ignore
	}

	// Send prompt if needed
	if state.Prompt != nil && !userState.PromptSent {
		return true, m.sendPrompt(update, &userState, state, key)
	}

	// Handle the update
	if state.Handle == nil {
		return false, nil
	}

	nextState, err := state.Handle(update, &userState.Data)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			return true, nil // Stay in current state
		}
		return false, err
	}

	// Update state
	userState.CurrentState = nextState
	userState.PromptSent = false
	if err := m.storage.Set(key, userState); err != nil {
		return false, err
	}

	// End of flow or no state transition
	if nextState == "" || nextState == NopState {
		return true, nil // End of flow
	}

	// Handle transition to next state
	if next, exists := m.states[nextState]; exists && next.Prompt != nil {
		return true, m.sendPrompt(update, &userState, next, key)
	}

	return true, nil
}

// sendPrompt is a helper function to send a prompt and update the state.
func (m *StateManager[S, U]) sendPrompt(update U, userState *UserState[S], state *State[S, U], key int64) error {
	if err := state.Prompt(update, &userState.Data); err != nil {
		return err
	}
	userState.PromptSent = true
	return m.storage.Set(key, *userState)
}
