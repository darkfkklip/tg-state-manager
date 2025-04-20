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
	bot, stateStorage := initializeServices()

	// Initialize state manager
	stateManager := initializeStateManager(stateStorage)

	// Register states
	registerStates(stateManager, bot)

	// Register handlers
	registerHandlers(bot, stateStorage, stateManager)

	fmt.Println("Bot started successfully.")
	bot.Start()
}

// initializeServices initializes and returns the bot and storage
func initializeServices() (*tele.Bot, tgsm.StateStorage[UserData]) {
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
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	return bot, tgsm.NewRedisStorage[UserData](redisClient, "user")
}

// initializeStateManager creates and returns the state manager
func initializeStateManager(storage tgsm.StateStorage[UserData]) *tgsm.StateManager[UserData, tele.Update] {
	return tgsm.NewStateManager(storage, func(u tele.Update) int64 {
		return u.Message.Chat.ID
	})
}

// registerStates adds all states to the state manager
func registerStates(stateManager *tgsm.StateManager[UserData, tele.Update], bot *tele.Bot) {
	if err := stateManager.Add(
		NewFirstNameState(bot),
		NewLastNameState(bot),
		NewSkillsState(bot),
		NewAgeState(bot),
		NewCooperateTypeState(bot),
	); err != nil {
		log.Fatalf("Failed to register states: %v", err)
	}
	stateManager.SetInitialState("first_name")
}

// registerHandlers sets up all bot command handlers
func registerHandlers(bot *tele.Bot, stateStorage tgsm.StateStorage[UserData], stateManager *tgsm.StateManager[UserData, tele.Update]) {
	// Middleware for state handling
	bot.Use(createStateMiddleware(stateManager))

	// Command handlers
	bot.Handle("/start", createStartHandler(stateStorage))
	bot.Handle("/profile", createProfileHandler(stateStorage))
	bot.Handle(tele.OnText, createDefaultHandler())
}

// createStateMiddleware returns middleware for state handling
func createStateMiddleware(stateManager *tgsm.StateManager[UserData, tele.Update]) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			handled, err := stateManager.Handle(c.Update())
			if err != nil {
				return fmt.Errorf("state handling error: %w", err)
			}
			if handled {
				return nil
			}
			return next(c)
		}
	}
}

func createStartHandler(stateStorage tgsm.StateStorage[UserData]) tele.HandlerFunc {
	return func(c tele.Context) error {
		userState, exists, err := stateStorage.Get(c.Message().Chat.ID)
		if err != nil {
			return fmt.Errorf("failed to get user state: %w", err)
		}
		
		if !exists || userState.Data.FirstName == "" {
			return stateStorage.Set(c.Message().Chat.ID, tgsm.UserState[UserData]{CurrentState: "first_name"})
		}
		return c.Send("Welcome back! Your registration is already complete.")
	}
}

// createProfileHandler returns a handler for the /profile command
func createProfileHandler(stateStorage tgsm.StateStorage[UserData]) tele.HandlerFunc {
	return func(c tele.Context) error {
		userState, exists, err := stateStorage.Get(c.Message().Chat.ID)
		if err != nil {
			return fmt.Errorf("failed to get user state: %w", err)
		}
		if !exists || userState.Data.FirstName == "" {
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

// createDefaultHandler returns a handler for text messages
func createDefaultHandler() tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("Please use commands /start, /profile.")
	}
}
