package impl

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

type RegisterCommandsCommand struct {
}

func (RegisterCommandsCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "registercommands",
		Description:     database.HelpAdmin, // TODO: Register translation
		Aliases:         []string{"registercmds", "rcmds", "rcmd"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		Arguments: command.Arguments(
			command.NewOptionalArgument("global", "Register commands globally", interaction.OptionTypeBoolean, database.MessageInvalidArgument),
		),
	}
}

func (c RegisterCommandsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (RegisterCommandsCommand) Execute(ctx command.CommandContext, global *bool) {
	for _, cmd := range Commands {
		properties := cmd.Properties()

		if properties.MessageOnly {
			continue
		}

		option := buildOption(cmd)

		data := rest.CreateCommandData{
			Name:        option.Name,
			Description: option.Description,
			Options:     option.Options,
		}

		var err error
		if global != nil && *global {
			_, err = ctx.Worker().CreateGlobalCommand(ctx.Worker().BotId, data)
		} else {
			_, err = ctx.Worker().CreateGuildCommand(ctx.Worker().BotId, ctx.GuildId(), data)
		}

		if err != nil {
			ctx.ReplyRaw(utils.Red, "Error", fmt.Sprintf("An error occurred while creating command `%s`: ```%v```", properties.Name, err))
			ctx.Reject()
			return
		}

		fmt.Printf("Registered %s\n", properties.Name)
	}

	ctx.Accept()
}

func buildOption(cmd command.Command) interaction.ApplicationCommandOption {
	properties := cmd.Properties()

	// Required args must come before optional args
	var required []interaction.ApplicationCommandOption
	var optional []interaction.ApplicationCommandOption

	for _, child := range properties.Children {
		if child.Properties().MessageOnly {
			continue
		}

		option := buildOption(child)

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
		Description: i18n.GetMessage(database.English, properties.Description),
		Default:     false,
		Required:    false,
		Choices:     nil,
		Options:     options,
	}
}
