package command

import (
	translations "github.com/TicketsBot/database/translations"
	"github.com/rxdn/gdl/objects/interaction"
)

type Argument struct {
	Name                   string
	Description            string
	Type                   interaction.ApplicationCommandOptionType
	Required               bool
	InvalidMessage         translations.MessageId
	SlashCommandCompatible bool
}

func NewOptionalArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		SlashCommandCompatible: true,
	}
}

func NewRequiredArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		SlashCommandCompatible: true,
	}
}

func NewOptionalArgumentMessageOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		SlashCommandCompatible: false,
	}
}

func NewRequiredArgumentMessageOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		SlashCommandCompatible: false,
	}
}

func Arguments(argument ...Argument) []Argument {
	return argument
}
