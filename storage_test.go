package tgstatemanager_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tgsm "github.com/sudosz/tg-state-manager"
)

type TestData struct {
	Name  string
	Value int
}

type testStorageConfig struct {
	client     *redis.Client
	testPrefix string
}

func setupTestEnv(t *testing.T) *testStorageConfig {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Skipping Redis tests: %v", err)
		return nil
	}

	return &testStorageConfig{
		client:     client,
		testPrefix: "test-storage-" + uuid.New().String(),
	}
}

func cleanupTestEnv(t *testing.T, cfg *testStorageConfig) {
	if cfg == nil || cfg.client == nil {
		return
	}

	ctx := context.Background()
	const batchSize = 1000
	iter := cfg.client.Scan(ctx, 0, cfg.testPrefix+":*", batchSize).Iterator()
	keys := make([]string, 0, batchSize)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= batchSize {
			require.NoError(t, cfg.client.Del(ctx, keys...).Err())
			keys = keys[:0]
		}
	}

	if len(keys) > 0 {
		require.NoError(t, cfg.client.Del(ctx, keys...).Err())
	}
	require.NoError(t, iter.Err())
	cfg.client.Close()
}

func createStorageFactories(cfg *testStorageConfig) map[string]func() tgsm.StateStorage[TestData] {
	factories := map[string]func() tgsm.StateStorage[TestData]{
		"InMemory": func() tgsm.StateStorage[TestData] {
			return tgsm.NewInMemoryStorage[TestData]()
		},
	}

	if cfg != nil && cfg.client != nil {
		factories["Redis"] = func() tgsm.StateStorage[TestData] {
			return tgsm.NewRedisStorage[TestData](cfg.client, cfg.testPrefix)
		}
	}

	return factories
}

func TestStorageImplementations(t *testing.T) {
	cfg := setupTestEnv(t)
	defer cleanupTestEnv(t, cfg)

	factories := createStorageFactories(cfg)
	testCases := []struct {
		name string
		test func(*testing.T, tgsm.StateStorage[TestData])
	}{
		{"Basic operations", testStorageOperations},
		{"Concurrent access", testConcurrentAccess},
		{"Edge cases", testEdgeCases},
	}

	for implName, factory := range factories {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s_%s", implName, tc.name), func(t *testing.T) {
				tc.test(t, factory())
			})
		}
	}
}

func FuzzStorageOperations(f *testing.F) {
	storage := tgsm.NewInMemoryStorage[TestData]()

	// Add seed corpus
	f.Add(int64(1), "test", 42, true)
	f.Add(int64(-1), "", -42, false)
	f.Add(int64(0), "long_string_test", 0, true)

	f.Fuzz(func(t *testing.T, userID int64, name string, value int, promptSent bool) {
		state := tgsm.UserState[TestData]{
			CurrentState: fmt.Sprintf("state_%s", strings.ReplaceAll(name, "/", "_")), // Sanitize state name
			Data: TestData{
				Name:  name,
				Value: value,
			},
			PromptSent: promptSent,
		}

		// Test Set operation
		if err := storage.Set(userID, state); err != nil {
			t.Skip()
		}

		// Test Get operation
		retrieved, exists, err := storage.Get(userID)
		if err != nil {
			t.Errorf("unexpected error during Get: %v", err)
		}
		if !exists {
			t.Error("state should exist after Set")
		}

		// Compare fields individually since struct comparison is not supported
		if retrieved.CurrentState != state.CurrentState ||
			retrieved.Data.Name != state.Data.Name ||
			retrieved.Data.Value != state.Data.Value ||
			retrieved.PromptSent != state.PromptSent {
			t.Error("retrieved state doesn't match stored state")
		}

		// Verify state exists
		_, exists, err = storage.Get(userID)
		if err != nil {
			t.Errorf("unexpected error checking state: %v", err)
		}
		if !exists {
			t.Error("state should exist")
		}
	})
}

func testStorageOperations(t *testing.T, storage tgsm.StateStorage[TestData]) {
	testCases := []struct {
		name     string
		userID   int64
		state    tgsm.UserState[TestData]
		validate func(*testing.T, tgsm.UserState[TestData], tgsm.UserState[TestData])
	}{
		{
			name:   "Basic state",
			userID: rand.Int63(),
			state: tgsm.UserState[TestData]{
				CurrentState: "initial",
				Data: TestData{
					Name:  "Test User",
					Value: 42,
				},
				PromptSent: true,
			},
			validate: func(t *testing.T, expected, actual tgsm.UserState[TestData]) {
				assert.Equal(t, expected, actual)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, storage.Set(tc.userID, tc.state))
			retrieved, exists, err := storage.Get(tc.userID)
			require.NoError(t, err)
			assert.True(t, exists)
			tc.validate(t, tc.state, retrieved)
		})
	}
}

func testConcurrentAccess(t *testing.T, storage tgsm.StateStorage[TestData]) {
	const numGoroutines = 50
	const numOperations = 20

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i := range numGoroutines {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := range numOperations {
				select {
				case <-ctx.Done():
					errors <- ctx.Err()
					return
				default:
					if err := performRandomOperation(storage, workerID, j); err != nil {
						errors <- err
					}
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		assert.NoError(t, err)
	}
}

func testEdgeCases(t *testing.T, storage tgsm.StateStorage[TestData]) {
	testCases := []struct {
		name     string
		userID   int64
		state    tgsm.UserState[TestData]
		validate func(*testing.T, tgsm.StateStorage[TestData], int64)
	}{
		{
			name:   "Zero UserID",
			userID: 0,
			state:  generateRandomState(),
			validate: func(t *testing.T, s tgsm.StateStorage[TestData], id int64) {
				retrieved, exists, err := s.Get(id)
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.NotEmpty(t, retrieved.CurrentState)
			},
		},
		{
			name:   "Empty State",
			userID: rand.Int63(),
			state:  tgsm.UserState[TestData]{},
			validate: func(t *testing.T, s tgsm.StateStorage[TestData], id int64) {
				retrieved, exists, err := s.Get(id)
				assert.NoError(t, err)
				assert.True(t, exists)
				assert.Empty(t, retrieved.CurrentState)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, storage.Set(tc.userID, tc.state))
			tc.validate(t, storage, tc.userID)
		})
	}
}

func performRandomOperation(storage tgsm.StateStorage[TestData], workerID, opID int) error {
	userID := rand.Int63()
	state := generateRandomState()
	state.Data.Name = fmt.Sprintf("Worker_%d_Op_%d", workerID, opID)

	if err := storage.Set(userID, state); err != nil {
		return fmt.Errorf("set error: %w", err)
	}

	retrieved, exists, err := storage.Get(userID)
	if err != nil {
		return fmt.Errorf("get error: %w", err)
	}

	if !exists || retrieved.Data.Name != state.Data.Name {
		return fmt.Errorf("state mismatch for worker %d operation %d", workerID, opID)
	}

	return nil
}

func generateRandomState() tgsm.UserState[TestData] {
	return tgsm.UserState[TestData]{
		CurrentState: fmt.Sprintf("state_%s", uuid.New().String()),
		Data: TestData{
			Name:  fmt.Sprintf("User_%s", uuid.New().String()),
			Value: rand.Intn(1000),
		},
		PromptSent: rand.Float32() < 0.5,
	}
}
