package command

import (
	translations "github.com/TicketsBot/database/translations"
	"github.com/rxdn/gdl/objects/interaction"
)

type Argument struct {
	Name           string
	Description    string
	Type           interaction.ApplicationCommandOptionType
	Required       bool
	InvalidMessage translations.MessageId
}

func NewArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, required bool, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:           name,
		Description:    description,
		Type:           argumentType,
		Required:       required,
		InvalidMessage: invalidMessage,
	}
}

func Arguments(argument ...Argument) []Argument {
	return argument
}
