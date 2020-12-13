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
	}
}

func (c RegisterCommandsCommand) GetExecutor() interface{} {
	return c.Execute
}

func (RegisterCommandsCommand) Execute(ctx command.CommandContext) {
	for _, cmd := range Commands {
		properties := cmd.Properties()

		//option := buildOption(cmd)

		data := rest.CreateCommandData{
			Name:        properties.Name,
			Description: i18n.GetMessage(database.English, properties.Description),
			Options:     make([]interaction.ApplicationCommandOption, 0),
		}

		// TODO: Make global
		if _, err := ctx.Worker.CreateGuildCommand(ctx.Worker.BotId, ctx.GuildId, data); err != nil {
			ctx.SendEmbedRaw(utils.Red, "Error", fmt.Sprintf("An error occurred while creating command `%s`: ```%v```", properties.Name, err))
			ctx.ReactWithCross()
			return
		}
	}

	ctx.ReactWithCheck()
}

func buildOption(cmd command.Command) interaction.ApplicationCommandOption {
	properties := cmd.Properties()

	var options []interaction.ApplicationCommandOption
	for _, child := range properties.Children {
		options = append(options, buildOption(child))
	}

	for _, argument := range properties.Arguments {
		options = append(options, interaction.ApplicationCommandOption{
			Type:        argument.Type,
			Name:        argument.Name,
			Description: argument.Description,
			Default:     false,
			Required:    argument.Required,
			Choices:     nil,
			Options:     nil,
		})
	}

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
