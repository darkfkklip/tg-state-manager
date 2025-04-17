package tgstatemanager

import (
	"errors"
)

// ValidationError indicates a validation failure, keeping the current state.
var ValidationError = errors.New("validation error")

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
	states  map[string]*State[S, U]
	storage StateStorage[S]
	keyFunc func(update U) int64
}

// NewStateManager creates a new StateManager.
func NewStateManager[S, U any](storage StateStorage[S], keyFunc func(update U) int64) *StateManager[S, U] {
	return &StateManager[S, U]{
		states:  make(map[string]*State[S, U]),
		storage: storage,
		keyFunc: keyFunc,
	}
}

// Append adds states to the manager.
func (m *StateManager[S, U]) Append(states ...*State[S, U]) {
	for _, state := range states {
		m.states[state.Name] = state
	}
}

// Handle processes an update, managing state transitions.
func (m *StateManager[S, U]) Handle(update U) (bool, error) {
	key := m.keyFunc(update)
	userState, exists, err := m.storage.Get(key)
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil // No state, ignore
	}

	state, stateExists := m.states[userState.CurrentState]
	if !stateExists {
		return false, nil // Invalid state, ignore
	}

	// Send prompt if needed
	if state.Prompt != nil && !userState.PromptSent {
		if err := m.sendPrompt(update, &userState, state, key); err != nil {
			return false, err
		}
		return true, nil
	}

	// Handle the update
	if state.Handle == nil {
		return false, nil
	}

	nextState, err := state.Handle(update, &userState.Data)
	if err != nil {
		if errors.Is(err, ValidationError) {
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

	if nextState == "" {
		return true, nil // End of flow
	}

	// Handle transition to next state
	if nextState != NopState {
		if nextState, exists := m.states[nextState]; exists && nextState.Prompt != nil {
			if err := m.sendPrompt(update, &userState, nextState, key); err != nil {
				return false, err
			}
			return true, nil
		}
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
