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
	MessageCompatible      bool
	SlashCommandCompatible bool
}

func NewOptionalArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      true,
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
		MessageCompatible:      true,
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
		MessageCompatible:      true,
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
		MessageCompatible:      true,
		SlashCommandCompatible: false,
	}
}

func NewOptionalArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      false,
		SlashCommandCompatible: true,
	}
}

func NewRequiredArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage translations.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      false,
		SlashCommandCompatible: true,
	}
}

func Arguments(argument ...Argument) []Argument {
	return argument
}
