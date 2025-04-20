# Telegram State Manager

![Go Version](https://img.shields.io/badge/Go-%3E%3D1.18-00ADD8)
![License](https://img.shields.io/github/license/sudosz/tg-state-manager)
[![GoDoc](https://godoc.org/github.com/sudosz/tg-state-manager?status.svg)](https://pkg.go.dev/github.com/sudosz/tg-state-manager)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/sudosz/tg-state-manager/actions)
[![Release](https://img.shields.io/github/v/release/sudosz/tg-state-manager)](https://github.com/sudosz/tg-state-manager/releases)

Telegram State Manager (`tg-state-manager`) is a lightweight, type-safe state management library for Telegram bots in Go. It simplifies multi-step conversations with a clean API, flexible storage options (in-memory or Redis), and compatibility with any Telegram bot framework. Whether you're building a small bot or a complex application, this library keeps state handling straightforward and intuitive.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Defining States](#defining-states)
  - [Configuring Storage](#configuring-storage)
  - [Setting Up the State Manager](#setting-up-the-state-manager)
  - [Handling Updates](#handling-updates)
  - [Starting a Conversation](#starting-a-conversation)
  - [Adding States](#adding-states)
  - [Best Practices](#best-practices)
- [Examples](#examples)
  - [Basic Examples](#basic-examples)
  	- [Telebot](#telebot)
  	- [Telego](#telego)
  	- [Telegram Bot API](#telegram-bot-api)
  - [Advanced Examples](#advanced-examples)
	- [Creating Reusable State Patterns](#creating-reusable-state-patterns)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [License](#license)
- [Contributors](#contributors)

## Features

- **Easy to Use**: Define states with simple `Prompt` and `Handle` functions.
- **Type-Safe**: Leverages Go generics for safety (requires Go 1.18+).
- **Flexible Storage**: Use in-memory or Redis backends.
- **Thread-Safe**: Built-in concurrency support for in-memory storage.
- **Framework-Agnostic**: Works with any Telegram bot library (e.g., [telebot](https://github.com/tucnak/telebot), [telego](https://github.com/mymmrac/telego), [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)).
- **Lightweight**: Minimal dependencies for quick integration.

## Installation

Get the library with:

```bash
go get github.com/sudosz/tg-state-manager
```

For Redis storage (optional):

```bash
go get github.com/redis/go-redis/v9
```

## Quick Start

Here's a simple example showing how to use `tg-state-manager` to collect a user's name and age. This uses the `telebot` framework, but the library works with any Telegram bot framework.

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

// UserData holds the information collected during the conversation
type UserData struct {
	Name string
	Age  int
}

func main() {
	// Initialize bot
	bot, err := tele.NewBot(tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Setup state manager with in-memory storage
	storage := tgsm.NewInMemoryStorage[UserData]()
	sm := tgsm.NewStateManager(storage, func(u tele.Update) int64 {
		return u.Message.Chat.ID
	})

	sm.SetInitialState("ask_name")
	err = sm.Add(
		&tgsm.State[UserData, tele.Update]{
			Name: "ask_name",
			Prompt: func(u tele.Update, data *UserData) error {
				_, err := bot.Send(u.Message.Chat, "What's your name?")
				return err
			},
			Handle: func(u tele.Update, data *UserData) (string, error) {
				if u.Message.Text == "" {
					return "", tgsm.ErrValidation
				}
				data.Name = u.Message.Text
				return "ask_age", nil
			},
		},
		&tgsm.State[UserData, tele.Update]{
			Name: "ask_age",
			Prompt: func(u tele.Update, data *UserData) error {
				_, err := bot.Send(u.Message.Chat, "How old are you?")
				return err
			},
			Handle: func(u tele.Update, data *UserData) (string, error) {
				age, err := strconv.Atoi(u.Message.Text)
				if err != nil || age < 0 {
					return "", tgsm.ErrValidation
				}
				data.Age = age
				_, err = bot.Send(u.Message.Chat, "All set!")
				return "", err
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// State management middleware
	bot.Use(func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			handled, err := sm.Handle(c.Update())
			if err != nil {
				return err
			}
			if !handled {
				return next(c)
			}
			return nil
		}
	})

	// Start command handler
	bot.Handle("/start", func(c tele.Context) error {
		userData, _, err := storage.Get(c.Chat().ID)
		if err != nil {
			return err
		}
		_, err = bot.Send(c.Message().Chat, fmt.Sprintf("Hello, %s!", userData.Data.Name))
		return err
	})

	// Fallback handler (required due to middleware)
	bot.Handle(tele.OnText, func(c tele.Context) error {
		return nil // Do nothing for unhandled text messages
	})

	bot.Start()
}
```

**Why the Fallback Handler?**  
Since we use middleware to process state manager updates, a fallback handler (like `tele.OnText`) is required to catch any updates the state manager doesn't handle. This prevents the bot from silently ignoring messages.

## Usage

### Defining States

Create states with the `State[S, U]` struct:
- `S`: Your custom data type (e.g., `UserData`).
- `U`: The update type from your framework.
- `Name`: A unique state identifier.
- `Prompt`: Sends a message when entering the state (optional).
- `Handle`: Processes input and returns the next state (or `""` to end).

```go
state := &tgsm.State[UserData, UpdateType]{
	Name: "example",
	Prompt: func(u UpdateType, data *UserData) error {
		// Send a message
		return nil
	},
	Handle: func(u UpdateType, data *UserData) (string, error) {
		// Process input, update data, return next state
		return "next_state", nil
	},
}
```

### Configuring Storage

Pick a storage backend:
- **In-Memory** (thread-safe):
  ```go
  storage := tgsm.NewInMemoryStorage[UserData]()
  ```
- **Redis** (persistent):
  ```go
  client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
  storage := tgsm.NewRedisStorage[UserData](client, "bot-prefix")
  ```

### Setting Up the State Manager

Initialize with storage and a key function:
```go
sm := tgsm.NewStateManager[UserData, UpdateType](storage, func(u UpdateType) int64 {
	return u.ChatID // Extract chat/user ID
})
```

### Handling Updates

Process updates with middleware or a handler:
```go
// Example middleware
bot.Use(func(next HandlerFunc) HandlerFunc {
	return func(c Context) error {
		handled, err := sm.Handle(c.Update())
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

Set an initial state:
```go
sm.SetInitialState("first_state")
```

### Adding States

Add states with validation to prevent duplicates:

```go
err := sm.Add(state1, state2, state3)
if err != nil {
   log.Fatalf("Error adding states: %v", err)
}
```

**Note**: The `Add` function returns an error if duplicate state names are detected to prevent accidental overwrites. This ensures state integrity by allowing developers to handle duplicates explicitly (e.g., logging or skipping them) rather than silently overwriting existing states, which could lead to bugs.

### Best Practices

1. **Group Related States**: Keep states for a specific flow together.
2. **Use Descriptive Names**: Name states clearly to understand their purpose.
3. **Handle Edge Cases**: Consider what happens when users send unexpected inputs.
4. **Provide Clear Feedback**: Always inform users about validation errors.
5. **End States**: Use empty string or `tgsm.NopState` to indicate the end of a conversation flow.

## Examples

### Basic Examples
Here's a simple examples of using the library with `telebot`, `telego` and `telegram-bot-api`:

#### Telebot

```go
package main

import (
	"log"
	"os"
	"time"

	tgsm "github.com/sudosz/tg-state-manager"
	tele "gopkg.in/telebot.v4"
)

type UserData struct {
	Name string
}

func main() {
	bot, err := tele.NewBot(tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	storage := tgsm.NewInMemoryStorage[UserData]()
	sm := tgsm.NewStateManager(storage, func(u tele.Update) int64 {
		return u.Message.Chat.ID
	})

	sm.SetInitialState("ask_name")

	err = sm.Add(&tgsm.State[UserData, tele.Update]{
		Name: "ask_name",
		Prompt: func(u tele.Update, data *UserData) error {
			_, err := bot.Send(u.Message.Chat, "What's your name?")
			return err
		},
		Handle: func(u tele.Update, data *UserData) (string, error) {
			data.Name = u.Message.Text
			_, err := bot.Send(u.Message.Chat, "Got it!")
			return "", err
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	bot.Handle(tele.OnText, func(c tele.Context) error {
		_, err := sm.Handle(c.Update())
		return err
	})

	bot.Start()
}
```

#### Telego

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/mymmrac/telego"
	tgsm "github.com/sudosz/tg-state-manager"
)

type UserData struct {
	Name string
}

func main() {
	ctx := context.Background()
	bot, err := telego.NewBot(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	storage := tgsm.NewInMemoryStorage[UserData]()
	sm := tgsm.NewStateManager[UserData, telego.Update](storage, func(u telego.Update) int64 {
		return u.Message.Chat.ID
	})

	sm.SetInitialState("ask_name")

	err = sm.Add(&tgsm.State[UserData, telego.Update]{
		Name: "ask_name",
		Prompt: func(u telego.Update, data *UserData) error {
			_, err := bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: u.Message.Chat.ID},
				Text:   "What's your name?",
			})
			return err
		},
		Handle: func(u telego.Update, data *UserData) (string, error) {
			data.Name = u.Message.Text
			_, err := bot.SendMessage(ctx, &telego.SendMessageParams{
				ChatID: telego.ChatID{ID: u.Message.Chat.ID},
				Text:   "Got it!",
			})
			return "", err
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	for update := range updates {
		_, err := sm.Handle(telego.Update(update))
		if err != nil {
			log.Println(err)
		}
	}
}
```

#### Telegram Bot API

```go
package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgsm "github.com/sudosz/tg-state-manager"
)

type UserData struct {
	Name string
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	storage := tgsm.NewInMemoryStorage[UserData]()
	sm := tgsm.NewStateManager[UserData, tgbotapi.Update](storage, func(u tgbotapi.Update) int64 {
		return u.Message.Chat.ID
	})

	sm.SetInitialState("ask_name")

	err = sm.Add(&tgsm.State[UserData, tgbotapi.Update]{
		Name: "ask_name",
		Prompt: func(u tgbotapi.Update, data *UserData) error {
			msg := tgbotapi.NewMessage(u.Message.Chat.ID, "What's your name?")
			_, err := bot.Send(msg)
			return err
		},
		Handle: func(u tgbotapi.Update, data *UserData) (string, error) {
			data.Name = u.Message.Text
			msg := tgbotapi.NewMessage(u.Message.Chat.ID, "Got it!")
			_, err := bot.Send(msg)
			return "", err
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		_, err := sm.Handle(update)
		if err != nil {
			log.Println(err)
		}
	}
}
```

### Advanced Examples

#### Creating Reusable State Patterns

For complex bots, consider creating helper functions that generate states with common patterns: For example, In `telebot` we can do this:

```go
// Helper function to create states with standard validation
func createState(
    bot *tele.Bot,
    name string,
    promptMsg string,
    validator func(string) (bool, error),
    errorMsg string,
    nextState string,
    updateState func(string, *UserData),
) *tgsm.State[UserData, tele.Update] {
    return &tgsm.State[UserData, tele.Update]{
        Name: name,
        Prompt: func(u tele.Update, state *UserData) error {
            _, err := bot.Send(u.Message.Chat, promptMsg)
            return err
        },
        Handle: func(u tele.Update, state *UserData) (string, error) {
            text := u.Message.Text
            if valid, _ := validator(text); !valid {
                _, err := bot.Send(u.Message.Chat, errorMsg)
                if err != nil {
                    return "", err
                }
                return "", tgsm.ErrValidation
            }
            updateState(text, state)
            return nextState, nil
        },
    }
}

// Using the helper to create a specific state
nameState := createState(
    bot,
    "ask_name",
    "What's your name?",
    func(text string) (bool, error) {
        return len(text) >= 2 && len(text) <= 50, nil
    },
    "Please enter a valid name (2-50 characters).",
    "ask_age",
    func(text string, data *UserData) {
        data.Name = text
    },
)
```

For more advanced examples, refer to the [examples](examples) directory.

## API Reference

| Name | Description |
| --- | --- |
| `StateManager[S, U any]` | Manages states and transitions. |
| `NewStateManager(storage, keyFunc)` | Creates a new state manager. |
| `Add(states ...*State[S, U]) error` | Adds states to the manager with duplicate checking. |
| `Handle(update U) (bool, error)` | Processes an update. |
| `State[S, U any]` | Defines a state with `Prompt`/`Handle`. |
| `UserState[S any]` | Holds current state and data. |
| `StateStorage[S any]` | Storage interface (`Get`, `Set`). |
| `NewInMemoryStorage[S any]()` | Creates in-memory storage. |
| `NewRedisStorage[S any](client, prefix)` | Creates Redis storage. |

## Contributing

We'd love your help! To contribute:
1. Fork the repo on [GitHub](https://github.com/sudosz/tg-state-manager).
2. Create a branch (`git checkout -b my-feature`).
3. Commit your changes (`git commit -m "Add feature"`).
4. Push (`git push origin my-feature`).
5. Open a pull request.

Please follow Go standards and add tests if possible. Report bugs or suggest ideas in the [issues](https://github.com/sudosz/tg-state-manager/issues) section.

## License

Licensed under the MIT License. See [LICENSE](LICENSE) for details.

## Contributors

Thanks to everyone who's helped improve `tg-state-manager`!

<a href="https://github.com/sudosz/tg-state-manager/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=sudosz/tg-state-manager" />
</a>

---

Enjoying `tg-state-manager`? Give it a ⭐ on [GitHub](https://github.com/sudosz/tg-state-manager)!
