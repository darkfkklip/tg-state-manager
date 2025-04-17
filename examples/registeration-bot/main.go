package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	tgsm "github.com/sudosz/tg-state-manager"
	tele "gopkg.in/telebot.v4"
)

// UserData holds the state data for a user.
type UserData struct {
	FirstName     string
	LastName      string
	Skills        string
	Age           int
	CooperateType string
}

func main() {
	// Initialize bot and Redis client
	bot, storage := setupServices()

	// Initialize state manager
	stateManager := setupStateManager(storage)

	// Register states
	registerStates(stateManager, bot)

	// Register handlers
	registerHandlers(bot, storage, stateManager)

	fmt.Println("Starting bot...")
	bot.Start()
}

// setupServices initializes and returns the bot and storage
func setupServices() (*tele.Bot, tgsm.StateStorage[UserData]) {
	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Initialize bot
	bot, err := tele.NewBot(tele.Settings{
		Token:   os.Getenv("TOKEN"),
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		Verbose: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialize storage
	storage := tgsm.NewRedisStorage[UserData](redisClient, "user")
	return bot, storage
}

// setupStateManager creates and returns the state manager
func setupStateManager(storage tgsm.StateStorage[UserData]) *tgsm.StateManager[UserData, tele.Update] {
	return tgsm.NewStateManager(storage, func(u tele.Update) int64 {
		return u.Message.Chat.ID
	})
}

// registerStates adds all states to the state manager
func registerStates(m *tgsm.StateManager[UserData, tele.Update], bot *tele.Bot) {
	m.Append(
		NewFirstNameState(bot),
		NewLastNameState(bot),
		NewSkillsState(bot),
		NewAgeState(bot),
		NewCooperateTypeState(bot),
	)
}

// registerHandlers sets up all bot command handlers
func registerHandlers(bot *tele.Bot, storage tgsm.StateStorage[UserData], m *tgsm.StateManager[UserData, tele.Update]) {
	// Middleware for state handling
	bot.Use(createStateMiddleware(m))

	// Command handlers
	bot.Handle("/start", handleStart(storage))
	bot.Handle("/profile", handleProfile(storage))
	bot.Handle(tele.OnText, handleDefaultText())
}

// createStateMiddleware returns middleware for state handling
func createStateMiddleware(m *tgsm.StateManager[UserData, tele.Update]) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			handled, err := m.Handle(c.Update())
			if err != nil {
				return err
			}
			if !handled {
				return next(c)
			}
			return nil
		}
	}
}

// handleStart returns a handler for the /start command
func handleStart(storage tgsm.StateStorage[UserData]) tele.HandlerFunc {
	return func(c tele.Context) error {
		userState, exists, err := storage.Get(c.Message().Chat.ID)
		if err != nil || !exists || userState.Data.FirstName == "" {
			newState := tgsm.UserState[UserData]{CurrentState: "first_name"}
			return storage.Set(c.Message().Chat.ID, newState)
		}
		return c.Send("Welcome! Let's get started.")
	}
}

// handleProfile returns a handler for the /profile command
func handleProfile(storage tgsm.StateStorage[UserData]) tele.HandlerFunc {
	return func(c tele.Context) error {
		userState, ok, err := storage.Get(c.Message().Chat.ID)
		if err != nil {
			return err
		}
		if !ok || userState.Data.FirstName == "" {
			return c.Send("No profile data. Use /start to register.")
		}
		return c.Send(fmt.Sprintf(
			"Profile:\nName: %s %s\nSkills: %s\nAge: %d\nCooperate Type: %s",
			userState.Data.FirstName,
			userState.Data.LastName,
			userState.Data.Skills,
			userState.Data.Age,
			userState.Data.CooperateType,
		))
	}
}

// handleDefaultText returns a handler for text messages
func handleDefaultText() tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("Please use commands /start, /profile.")
	}
}
