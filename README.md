# tg-state-manager

A lightweight, type-safe, and easy-to-use state management library for Telegram bots in Go. Built with generics, `tg-state-manager` simplifies multi-step conversations by providing a clean API for defining states, handling transitions, and storing data with in-memory or Redis backends. Perfect for developers building interactive Telegram bots with frameworks like [telebot](https://github.com/tucnak/telebot).

[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](https://opensource.org/licenses/MIT)
[![GitHub Issues](https://img.shields.io/github/issues/sudosz/tg-state-manager)](https://github.com/sudosz/tg-state-manager/issues)

## Features

- **Simple API**: Define states with intuitive `Prompt` and `Handle` functions.
- **Type-Safe**: Uses Go generics for flexible, compile-time-checked state and update types.
- **Flexible Storage**: Supports thread-safe in-memory and persistent Redis backends.
- **Concurrent**: In-memory storage is safe for concurrent access with `sync.RWMutex`.
- **Lightweight**: Minimal dependencies, optimized for performance.
- **Framework-Agnostic**: Works with any Telegram bot framework (e.g., telebot, go-telegram-bot-api).

## Installation

Install the package using:

```bash
go get github.com/sudosz/tg-state-manager
```

For Redis storage, install the optional dependency:

```bash
go get github.com/redis/go-redis/v9
```

## Quick Start

Below is a complete example of a Telegram bot that collects a user’s name and age using `tg-state-manager` with the [telebot](https://github.com/tucnak/telebot) framework.

```go
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	tgsm "github.com/sudosz/tg-state-manager"
	tele "gopkg.in/telebot.v4"
)

// UserData holds state data
type UserData struct {
	Name string
	Age  int
}

func main() {
	// Initialize bot
	b, err := tele.NewBot(tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	storage := tgsm.NewInMemoryStorage[UserData]()

	// Create state manager
	m := tgsm.NewStateManager(storage, func(u tele.Update) int64 {
		return u.Message.Chat.ID
	})

	// Define states
	m.Append(
		&tgsm.State[UserData, tele.Update]{
			Name: "ask_name",
			Prompt: func(u tele.Update, state *UserData) error {
				_, err := b.Send(u.Message.Chat, "Please enter your name.")
				return err
			},
			Handle: func(u tele.Update, state *UserData) (string, error) {
				if u.Message.Text == "" {
					return "", tgsm.ValidationError
				}
				state.Name = u.Message.Text
				return "ask_age", nil
			},
		},
		&tgsm.State[UserData, tele.Update]{
			Name: "ask_age",
			Prompt: func(u tele.Update, state *UserData) error {
				_, err := b.Send(u.Message.Chat, "Please enter your age.")
				return err
			},
			Handle: func(u tele.Update, state *UserData) (string, error) {
				age, err := strconv.Atoi(u.Message.Text)
				if err != nil || age < 0 {
					return "", tgsm.ValidationError
				}
				state.Age = age
				_, err = b.Send(u.Message.Chat, "Registration complete!")
				return "", err
			},
		},
	)

	// Middleware for state handling
	b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
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
	})

	// Start command
	b.Handle("/start", func(c tele.Context) error {
		userState := tgsm.UserState[UserData]{CurrentState: "ask_name"}
		if err := storage.Set(c.Chat().ID, userState); err != nil {
			return err
		}
		_, err := m.Handle(c.Update())
		return err
	})

	// Profile command
	b.Handle("/profile", func(c tele.Context) error {
		userState, ok, err := storage.Get(c.Chat().ID)
		if err != nil {
			return err
		}
		if !ok || userState.Data.Name == "" {
			return c.Send("No profile data. Use /start to register.")
		}
		return c.Send(fmt.Sprintf("Name: %s\nAge: %d", userState.Data.Name, userState.Data.Age))
	})

	// Fallback for getting inputs
	b.Handle(tele.OnText, func(c tele.Context) error {
		return nil
	})

	log.Println("Bot running...")
	b.Start()
}
```

## Examples
For more examples, see the [examples](examples) directory.

## Usage

### Defining States
States are defined using the `State[S, U]` struct, where `S` is your state data type and `U` is the update type (e.g., `tele.Update` for telebot). Each state includes:

- `Name`: A unique identifier.
- `Prompt`: An optional function to send a message when entering the state.
- `Handle`: A function to process input and return the next state (or `""` to end).

Example state for collecting a name:

```go
var nameState = &tgsm.State[UserData, tele.Update]{
	Name: "ask_name",
	Prompt: func(u tele.Update, state *UserData) error {
		_, err := b.Send(u.Message.Chat, "Please enter your name.")
		return err
	},
	Handle: func(u tele.Update, state *UserData) (string, error) {
		if u.Message.Text == "" {
			return "", tgsm.ValidationError
		}
		state.Name = u.Message.Text
		return "next_state", nil
	},
}
```

### Configuring Storage
Choose between in-memory or Redis storage:

- **In-Memory**:
  ```go
  storage := tgsm.NewInMemoryStorage[UserData]()
  ```

- **Redis**:
  ```go
  client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
  storage := tgsm.NewRedisStorage[UserData](client, "bot-state")
  ```

### Setting Up the State Manager
Create a `StateManager` with a storage backend and a key function to extract user or chat IDs:

```go
m := tgsm.NewStateManager[UserData, tele.Update](storage, func(u tele.Update) int64 {
	return u.Message.Chat.ID
})
m.Append(nameState, ageState)
```

### Handling Updates
You can use a middleware to process updates through the state manager:
For example, with telebot:
```go
b.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
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
})
```

### Starting a Conversation
You can set an initial state to begin the conversation:
For example, with telebot:
```go
b.Handle("/start", func(c tele.Context) error {
	userState := tgsm.UserState[UserData]{CurrentState: "ask_name"}
	return storage.Set(c.Chat().ID, userState)
})
```

## API Reference

| Type/Method | Description |
|-------------|-------------|
| `StateManager[S, U any]` | Manages state transitions. |
| `NewStateManager(storage StateStorage[S], keyFunc func(update U) int64) *StateManager[S, U]` | Initializes a new state manager. |
| `Append(states ...*State[S, U])` | Registers states. |
| `Handle(update U) (bool, error)` | Processes an update, returning whether it was handled. |
| `State[S, U any]` | Defines a state with `Name`, `Prompt`, and `Handle`. |
| `UserState[S any]` | Stores `CurrentState`, `Data`, and `PromptSent`. |
| `StateStorage[S any]` | Interface for storage backends with `Get` and `Set`. |
| `InMemoryStorage[S any]` | Thread-safe in-memory storage. |
| `RedisStorage[S any]` | Persistent Redis storage. |

## Contributing

We welcome contributions to `tg-state-manager`! To contribute:

1. Fork the repository on [GitHub](https://github.com/sudosz/tg-state-manager).
2. Create a feature or bugfix branch (`git checkout -b feature/my-feature`).
3. Commit your changes (`git commit -m "Add my feature"`).
4. Push to your branch (`git push origin feature/my-feature`).
5. Open a pull request with a clear description of your changes.

Please ensure your code follows Go conventions and includes tests where applicable. Report issues or suggest features via the [issue tracker](https://github.com/sudosz/tg-state-manager/issues).

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT). See the [LICENSE](LICENSE) file for details.

## Support

For questions or support, open an issue on the [GitHub repository](https://github.com/sudosz/tg-state-manager) or join the discussion in the [Go Telegram Bot API community](https://github.com/go-telegram-bot-api/telegram-bot-api).
