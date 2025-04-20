package main

import (
	"regexp"
	"strconv"

	tgsm "github.com/sudosz/tg-state-manager"
	tele "gopkg.in/telebot.v4"
)

// createState is a helper function to create state handlers with common validation patterns
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

// NewFirstNameState creates a state for collecting the user's first name.
func NewFirstNameState(bot *tele.Bot) *tgsm.State[UserData, tele.Update] {
	return createState(
		bot,
		"first_name",
		"Please enter your first name (3-16 characters).",
		func(text string) (bool, error) {
			return regexp.MatchString(`^[\p{L}\s]{3,16}$`, text)
		},
		"Invalid first name. Please use 3-16 characters.",
		"last_name",
		func(text string, state *UserData) {
			state.FirstName = text
		},
	)
}

// NewLastNameState creates a state for collecting the user's last name.
func NewLastNameState(bot *tele.Bot) *tgsm.State[UserData, tele.Update] {
	return createState(
		bot,
		"last_name",
		"Please enter your last name (4-20 characters).",
		func(text string) (bool, error) {
			return regexp.MatchString(`^[\p{L}\s]{4,20}$`, text)
		},
		"Invalid last name. Please use 4-20 characters.",
		"skills",
		func(text string, state *UserData) {
			state.LastName = text
		},
	)
}

// NewSkillsState creates a state for collecting the user's skills.
func NewSkillsState(bot *tele.Bot) *tgsm.State[UserData, tele.Update] {
	return createState(
		bot,
		"skills",
		"Please list your skills (16-512 characters).",
		func(text string) (bool, error) {
			return regexp.MatchString(`^[\p{L}\p{N}\s\-\(\)\+\*\%\#\@\~\,\.\;\:]{16,512}$`, text)
		},
		"Invalid skills format. Please use 16-512 characters.",
		"age",
		func(text string, state *UserData) {
			state.Skills = text
		},
	)
}

// NewAgeState creates a state for collecting the user's age.
func NewAgeState(bot *tele.Bot) *tgsm.State[UserData, tele.Update] {
	return createState(
		bot,
		"age",
		"Please enter your age (between 10 and 60).",
		func(text string) (bool, error) {
			age, err := strconv.Atoi(text)
			return err == nil && age >= 10 && age <= 60, nil
		},
		"Invalid age. Please enter a number between 10 and 60.",
		"cooperate_type",
		func(text string, state *UserData) {
			age, _ := strconv.Atoi(text)
			state.Age = age
		},
	)
}

// NewCooperateTypeState creates a state for collecting the user's cooperation type.
func NewCooperateTypeState(bot *tele.Bot) *tgsm.State[UserData, tele.Update] {
	return &tgsm.State[UserData, tele.Update]{
		Name: "cooperate_type",
		Prompt: func(u tele.Update, state *UserData) error {
			_, err := bot.Send(u.Message.Chat, "Please select your preferred cooperation type:", &tele.ReplyMarkup{
				ReplyKeyboard: [][]tele.ReplyButton{
					{{Text: "Full-time"}, {Text: "Part-time"}, {Text: "Project-based"}},
				},
				ResizeKeyboard:  true,
				OneTimeKeyboard: true,
			})
			return err
		},
		Handle: func(u tele.Update, state *UserData) (string, error) {
			valid := map[string]bool{"Full-time": true, "Part-time": true, "Project-based": true}
			if !valid[u.Message.Text] {
				_, err := bot.Send(u.Message.Chat, "Invalid selection. Please choose Full-time, Part-time, or Project-based.")
				if err != nil {
					return "", err
				}
				return "", tgsm.ErrValidation
			}
			state.CooperateType = u.Message.Text
			_, err := bot.Send(u.Message.Chat, "Registration completed successfully! Use /profile to view your information.")
			return tgsm.NopState, err
		},
	}
}
