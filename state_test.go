package tgstatemanager_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tgsm "github.com/sudosz/tg-state-manager"
)

type (
	// MockUpdate simulates a Telegram update
	MockUpdate struct {
		ChatID int64
		Text   string
	}

	// UserProfile represents user data collected during conversation
	UserProfile struct {
		Name    string
		Age     int
		Country string
	}
)

func setupStateManager(t testing.TB, storage tgsm.StateStorage[UserProfile]) *tgsm.StateManager[UserProfile, MockUpdate] {
	sm := tgsm.NewStateManager[UserProfile, MockUpdate](storage, func(u MockUpdate) int64 { return u.ChatID })
	sm.SetInitialState("ask_name")

	states := []*tgsm.State[UserProfile, MockUpdate]{
		createNameState(),
		createAgeState(),
		createCountryState(),
	}

	require.NoError(t, sm.Add(states...))
	return sm
}

func createNameState() *tgsm.State[UserProfile, MockUpdate] {
	return &tgsm.State[UserProfile, MockUpdate]{
		Name:   "ask_name",
		Prompt: func(u MockUpdate, data *UserProfile) error { return nil },
		Handle: func(u MockUpdate, data *UserProfile) (string, error) {
			if u.Text == "" {
				return "", tgsm.ErrValidation
			}
			data.Name = u.Text
			return "ask_age", nil
		},
	}
}

func createAgeState() *tgsm.State[UserProfile, MockUpdate] {
	return &tgsm.State[UserProfile, MockUpdate]{
		Name:   "ask_age",
		Prompt: func(u MockUpdate, data *UserProfile) error { return nil },
		Handle: func(u MockUpdate, data *UserProfile) (string, error) {
			age, err := strconv.Atoi(u.Text)
			if err != nil || age < 0 {
				return "", tgsm.ErrValidation
			}
			data.Age = age
			return "ask_country", nil
		},
	}
}

func createCountryState() *tgsm.State[UserProfile, MockUpdate] {
	return &tgsm.State[UserProfile, MockUpdate]{
		Name:   "ask_country",
		Prompt: func(u MockUpdate, data *UserProfile) error { return nil },
		Handle: func(u MockUpdate, data *UserProfile) (string, error) {
			if u.Text == "" {
				return "", tgsm.ErrValidation
			}
			data.Country = u.Text
			return "", nil
		},
	}
}

func TestStateManagerE2E(t *testing.T) {
	t.Run("InMemoryStorage", func(t *testing.T) {
		runE2ETest(t, tgsm.NewInMemoryStorage[UserProfile]())
	})

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		t.Run("RedisStorage", func(t *testing.T) {
			client := setupRedisClient(t, redisURL)
			defer client.Close()

			storage := tgsm.NewRedisStorage[UserProfile](client, "test-e2e")

			cleanupRedisKeys(t, client, "test-e2e:*")
			runE2ETest(t, storage)
		})
	}
}

func setupRedisClient(t *testing.T, redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, client.Ping(ctx).Err())
	return client
}

func cleanupRedisKeys(t *testing.T, client *redis.Client, pattern string) {
	ctx := context.Background()
	keys, err := client.Keys(ctx, pattern).Result()
	require.NoError(t, err)
	if len(keys) > 0 {
		require.NoError(t, client.Del(ctx, keys...).Err())
	}
}

func runE2ETest(t *testing.T, storage tgsm.StateStorage[UserProfile]) {
	sm := setupStateManager(t, storage)
	chatID := int64(123456789)

	testCases := []struct {
		name        string
		input       string
		wantState   string
		wantHandled bool
		wantError   bool
	}{
		{"Initial prompt", "", "ask_name", true, false},
		{"Provide name", "John Doe", "ask_age", true, false},
		{"Invalid age", "not a number", "ask_age", true, false},
		{"Valid age", "30", "ask_country", true, false},
		{"Provide country", "United States", "", true, false},
		{"No more state", "extra message", "", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handled, err := sm.Handle(MockUpdate{ChatID: chatID, Text: tc.input})

			assertTestCase(t, tc, handled, err, chatID, storage)
		})
	}

	verifyFinalState(t, storage, chatID)
}

func assertTestCase(t *testing.T, tc struct {
	name        string
	input       string
	wantState   string
	wantHandled bool
	wantError   bool
}, handled bool, err error, chatID int64, storage tgsm.StateStorage[UserProfile],
) {
	if tc.wantError {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}
	assert.Equal(t, tc.wantHandled, handled)

	if tc.wantState != "" || tc.name == "No more state" {
		state, exists, err := storage.Get(chatID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, tc.wantState, state.CurrentState)
	}
}

func verifyFinalState(t *testing.T, storage tgsm.StateStorage[UserProfile], chatID int64) {
	state, exists, err := storage.Get(chatID)
	require.NoError(t, err)
	require.True(t, exists)
	assert.Equal(t, "John Doe", state.Data.Name)
	assert.Equal(t, 30, state.Data.Age)
	assert.Equal(t, "United States", state.Data.Country)
}

func FuzzStateManager(f *testing.F) {
	f.Add("John", "25", "USA")
	f.Add("", "-1", "")
	f.Add("Alice", "abc", "Canada")

	f.Fuzz(func(t *testing.T, name, age, country string) {
		storage := tgsm.NewInMemoryStorage[UserProfile]()
		sm := setupStateManager(t, storage)
		chatID := int64(1)

		runFuzzTest(t, sm, storage, chatID, name, age, country)
	})
}

func runFuzzTest(t *testing.T, sm *tgsm.StateManager[UserProfile, MockUpdate], storage tgsm.StateStorage[UserProfile], chatID int64, name, age, country string) {
	handled, err := sm.Handle(MockUpdate{ChatID: chatID, Text: ""})
	assert.NoError(t, err)
	assert.True(t, handled)

	handled, err = sm.Handle(MockUpdate{ChatID: chatID, Text: name})
	assert.NoError(t, err)
	assert.True(t, handled)

	state, exists, err := storage.Get(chatID)
	assert.NoError(t, err)
	handleNameState(t, state, exists, name)

	if exists && state.CurrentState == "ask_age" {
		handleAgeState(t, sm, storage, chatID, age, country)
	}
}

func handleNameState(t *testing.T, state tgsm.UserState[UserProfile], exists bool, name string) {
	if name == "" {
		assert.Equal(t, "ask_name", state.CurrentState)
	} else {
		assert.Equal(t, "ask_age", state.CurrentState)
		assert.Equal(t, name, state.Data.Name)
	}
}

func handleAgeState(t *testing.T, sm *tgsm.StateManager[UserProfile, MockUpdate], storage tgsm.StateStorage[UserProfile], chatID int64, age, country string) {
	handled, err := sm.Handle(MockUpdate{ChatID: chatID, Text: age})
	assert.NoError(t, err)
	assert.True(t, handled)

	state, exists, err := storage.Get(chatID)
	assert.NoError(t, err)
	assert.True(t, exists)

	ageNum, err := strconv.Atoi(age)
	if err != nil || ageNum < 0 {
		assert.Equal(t, "ask_age", state.CurrentState)
	} else {
		assert.Equal(t, "ask_country", state.CurrentState)
		assert.Equal(t, ageNum, state.Data.Age)
	}

	if state.CurrentState == "ask_country" {
		handleCountryState(t, sm, storage, chatID, country)
	}
}

func handleCountryState(t *testing.T, sm *tgsm.StateManager[UserProfile, MockUpdate], storage tgsm.StateStorage[UserProfile], chatID int64, country string) {
	handled, err := sm.Handle(MockUpdate{ChatID: chatID, Text: country})
	assert.NoError(t, err)
	assert.True(t, handled)

	state, exists, err := storage.Get(chatID)
	assert.NoError(t, err)
	assert.True(t, exists)

	if country == "" {
		assert.Equal(t, "ask_country", state.CurrentState)
	} else {
		assert.Equal(t, "", state.CurrentState)
		assert.Equal(t, country, state.Data.Country)
	}
}

func TestStateManagerDuplicateStates(t *testing.T) {
	storage := tgsm.NewInMemoryStorage[UserProfile]()
	sm := tgsm.NewStateManager[UserProfile, MockUpdate](storage, func(u MockUpdate) int64 {
		return u.ChatID
	})

	testState := &tgsm.State[UserProfile, MockUpdate]{
		Name:   "test_state",
		Prompt: func(u MockUpdate, data *UserProfile) error { return nil },
		Handle: func(u MockUpdate, data *UserProfile) (string, error) { return "", nil },
	}

	require.NoError(t, sm.Add(testState))
	assert.Error(t, sm.Add(testState), "Adding duplicate state should return an error")
}

func TestStateManagerValidation(t *testing.T) {
	storage := tgsm.NewInMemoryStorage[UserProfile]()
	sm := tgsm.NewStateManager[UserProfile, MockUpdate](storage, func(u MockUpdate) int64 {
		return u.ChatID
	})

	sm.SetInitialState("validation_test")
	require.NoError(t, sm.Add(&tgsm.State[UserProfile, MockUpdate]{
		Name:   "validation_test",
		Prompt: func(u MockUpdate, data *UserProfile) error { return nil },
		Handle: func(u MockUpdate, data *UserProfile) (string, error) {
			if u.Text != "valid" {
				return "", tgsm.ErrValidation
			}
			return "next_state", nil
		},
	}))

	chatID := int64(987654321)
	testCases := []struct {
		input     string
		wantState string
	}{
		{"", "validation_test"},
		{"invalid", "validation_test"},
		{"valid", "next_state"},
	}

	for _, tc := range testCases {
		handled, err := sm.Handle(MockUpdate{ChatID: chatID, Text: tc.input})
		assert.NoError(t, err)
		assert.True(t, handled)

		state, exists, err := storage.Get(chatID)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, tc.wantState, state.CurrentState)
	}
}
