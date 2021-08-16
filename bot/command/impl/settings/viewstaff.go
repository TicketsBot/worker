package settings

import (
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/i18n"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
)

type ViewStaffCommand struct {
}

func (ViewStaffCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "viewstaff",
		Description:     i18n.HelpViewStaff,
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
	}
}

func (c ViewStaffCommand) GetExecutor() interface{} {
	return c.Execute
}

func (ViewStaffCommand) Execute(ctx registry.CommandContext) {
	msgEmbed, _ := logic.BuildViewStaffMessage(ctx.GuildId(), ctx.Worker(), 0, ctx.ToErrorContext())

	res := registry.MessageResponse{
		Embeds: []*embed.Embed{msgEmbed},
		Flags:  message.SumFlags(message.FlagEphemeral),
		Components: []component.Component{
			component.BuildActionRow(
				component.BuildButton(component.Button{
					CustomId: "disabled",
					Style:    component.ButtonStylePrimary,
					Emoji: emoji.Emoji{
						Name: "◀️",
					},
					Disabled: true,
				}),
				component.BuildButton(component.Button{
					CustomId: "viewstaff_1",
					Style:    component.ButtonStylePrimary,
					Emoji: emoji.Emoji{
						Name: "▶️",
					},
					Disabled: false,
				}),
			),
		},
	}

	_, _ = ctx.ReplyWith(res)
}
