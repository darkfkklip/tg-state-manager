package tgstatemanager

import "sync"

// InMemoryStorage provides a thread-safe in-memory storage for user states.
type InMemoryStorage[S any] struct {
	states map[int64]UserState[S]
	mu     sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage instance.
func NewInMemoryStorage[S any]() *InMemoryStorage[S] {
	return &InMemoryStorage[S]{
		states: make(map[int64]UserState[S]),
	}
}

// Get retrieves the user state for a given ID.
func (s *InMemoryStorage[S]) Get(id int64) (UserState[S], bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	state, ok := s.states[id]
	return state, ok, nil
}

// Set stores the user state for a given ID.
func (s *InMemoryStorage[S]) Set(id int64, userState UserState[S]) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[id] = userState
	return nil
}