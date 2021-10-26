package command

import (
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type Argument struct {
	Name                   string
	Description            string
	Type                   interaction.ApplicationCommandOptionType
	Required               bool
	InvalidMessage         i18n.MessageId
	MessageCompatible      bool
	SlashCommandCompatible bool
	AutoCompleteHandler    AutoCompleteHandler
}

type AutoCompleteHandler func(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice

func NewOptionalArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      true,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    nil,
	}
}

func NewRequiredArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      true,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    nil,
	}
}

func NewOptionalAutocompleteableArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      true,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    autoCompleteHandler,
	}
}

func NewRequiredAutocompleteableArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      true,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    autoCompleteHandler,
	}
}

func NewOptionalArgumentMessageOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
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

func NewRequiredArgumentMessageOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
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

func NewOptionalArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
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

func NewRequiredArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      false,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    nil,
	}
}

func NewOptionalAutocompleteableArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               false,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      false,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    autoCompleteHandler,
	}
}

func NewRequiredAutocompleteableArgumentInteractionOnly(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                   name,
		Description:            description,
		Type:                   argumentType,
		Required:               true,
		InvalidMessage:         invalidMessage,
		MessageCompatible:      false,
		SlashCommandCompatible: true,
		AutoCompleteHandler:    autoCompleteHandler,
	}
}

func Arguments(argument ...Argument) []Argument {
	return argument
}
