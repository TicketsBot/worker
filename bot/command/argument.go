package command

import (
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
)

type Argument struct {
	Name                string
	Description         string
	Type                interaction.ApplicationCommandOptionType
	Required            bool
	InvalidMessage      i18n.MessageId
	AutoCompleteHandler AutoCompleteHandler
}

type AutoCompleteHandler func(data interaction.ApplicationCommandAutoCompleteInteraction, value string) []interaction.ApplicationCommandOptionChoice

func NewOptionalArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
	return Argument{
		Name:                name,
		Description:         description,
		Type:                argumentType,
		Required:            false,
		InvalidMessage:      invalidMessage,
		AutoCompleteHandler: nil,
	}
}

func NewRequiredArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId) Argument {
	return Argument{
		Name:                name,
		Description:         description,
		Type:                argumentType,
		Required:            true,
		InvalidMessage:      invalidMessage,
		AutoCompleteHandler: nil,
	}
}

func NewOptionalAutocompleteableArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                name,
		Description:         description,
		Type:                argumentType,
		Required:            false,
		InvalidMessage:      invalidMessage,
		AutoCompleteHandler: autoCompleteHandler,
	}
}

func NewRequiredAutocompleteableArgument(name, description string, argumentType interaction.ApplicationCommandOptionType, invalidMessage i18n.MessageId, autoCompleteHandler AutoCompleteHandler) Argument {
	return Argument{
		Name:                name,
		Description:         description,
		Type:                argumentType,
		Required:            true,
		InvalidMessage:      invalidMessage,
		AutoCompleteHandler: autoCompleteHandler,
	}
}

func Arguments(argument ...Argument) []Argument {
	return argument
}
