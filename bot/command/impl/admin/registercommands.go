package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type RegisterCommandsCommand struct {
	Registry registry.Registry
}

func (RegisterCommandsCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "registercommands",
		Description:     i18n.HelpAdmin, // TODO: Register translation
		Aliases:         []string{"registercmds", "rcmds", "rcmd"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewOptionalArgument("global", "Register commands globally", interaction.OptionTypeBoolean, i18n.MessageInvalidArgument),
		),
	}
}

func (c RegisterCommandsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c RegisterCommandsCommand) Execute(ctx registry.CommandContext, global *bool) {
	for _, cmd := range c.Registry {
		properties := cmd.Properties()

		if properties.MessageOnly {
			continue
		}

		option := BuildOption(cmd)

		data := rest.CreateCommandData{
			Name:        option.Name,
			Description: option.Description,
			Options:     option.Options,
			Type:        properties.Type,
		}

		var err error
		if global != nil && *global {
			_, err = ctx.Worker().CreateGlobalCommand(ctx.Worker().BotId, data)
		} else {
			_, err = ctx.Worker().CreateGuildCommand(ctx.Worker().BotId, ctx.GuildId(), data)
		}

		if err != nil {
			ctx.ReplyRaw(customisation.Red, ctx.GetMessage(i18n.Error), fmt.Sprintf("An error occurred while creating command `%s`: ```%v```", properties.Name, err))
			ctx.Reject()
			return
		}

		fmt.Printf("Registered %s\n", properties.Name)
	}

	ctx.Accept()
}

func BuildOption(cmd registry.Command) interaction.ApplicationCommandOption {
	properties := cmd.Properties()

	// Required args must come before optional args
	var required []interaction.ApplicationCommandOption
	var optional []interaction.ApplicationCommandOption

	for _, child := range properties.Children {
		if child.Properties().MessageOnly {
			continue
		}

		option := BuildOption(child)

		if option.Required {
			required = append(required, option)
		} else {
			optional = append(optional, option)
		}
	}

	for _, argument := range properties.Arguments {
		if !argument.SlashCommandCompatible {
			continue
		}

		option := interaction.ApplicationCommandOption{
			Type:        argument.Type,
			Name:        argument.Name,
			Description: argument.Description,
			Default:     false,
			Required:    argument.Required,
			Choices:     nil,
			Options:     nil,
		}

		if option.Required {
			required = append(required, option)
		} else {
			optional = append(optional, option)
		}
	}

	options := append(required, optional...)

	return interaction.ApplicationCommandOption{
		Type:        interaction.OptionTypeSubCommand,
		Name:        properties.Name,
		Description: i18n.GetMessage(i18n.English, properties.Description),
		Default:     false,
		Required:    false,
		Choices:     nil,
		Options:     options,
	}
}
