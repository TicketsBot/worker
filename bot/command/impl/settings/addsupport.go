package settings

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/context"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
)

type AddSupportCommand struct {
	Registry registry.Registry
}

func (AddSupportCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "addsupport",
		Description:     i18n.HelpAddSupport,
		Type:            interaction.ApplicationCommandTypeChatInput,
		Aliases:         []string{"addsuport"},
		PermissionLevel: permcache.Admin,
		Category:        command.Settings,
		Arguments: command.Arguments(
			command.NewRequiredArgument("user_or_role", "User or role to apply the support representative permission to", interaction.OptionTypeMentionable, i18n.MessageAddSupportNoMembers),
		),
	}
}

func (c AddSupportCommand) GetExecutor() interface{} {
	return c.Execute
}

func (c AddSupportCommand) Execute(ctx registry.CommandContext, id uint64) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`/addsupport @User`\n`/addsupport @Role`",
		Inline: false,
	}

	mentionableType, valid := context.DetermineMentionableType(ctx, id)
	if !valid {
		ctx.ReplyWithFields(customisation.Red, i18n.Error, i18n.MessageAddSupportNoMembers, utils.ToSlice(usageEmbed))
		ctx.Reject()
		return
	}

	if mentionableType == context.MentionableTypeUser {
		// Guild owner doesn't need to be added
		guild, err := ctx.Guild()
		if err != nil {
			ctx.HandleError(err)
			return
		}

		if guild.OwnerId == id {
			ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOwnerIsAlreadyAdmin)
			return
		}

		if err := dbclient.Client.Permissions.AddSupport(ctx.GuildId(), id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	} else if mentionableType == context.MentionableTypeRole {
		if err := dbclient.Client.RolePermissions.AddSupport(ctx.GuildId(), id); err != nil {
			ctx.HandleError(err)
			return
		}

		if err := utils.ToRetriever(ctx.Worker()).Cache().SetCachedPermissionLevel(ctx.GuildId(), id, permcache.Support); err != nil {
			ctx.HandleError(err)
			return
		}
	} else {
		ctx.HandleError(fmt.Errorf("infallible"))
		return
	}

	ctx.Reply(customisation.Green, i18n.TitleAddSupport, i18n.MessageAddSupportSuccess)
	ctx.Accept()

	updateChannelPermissions(ctx, id, mentionableType)
}

func updateChannelPermissions(ctx registry.CommandContext, id uint64, mentionableType context.MentionableType) {
	openTickets, err := dbclient.Client.Tickets.GetGuildOpenTickets(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil {
			continue
		}

		ch, err := ctx.Worker().GetChannel(*ticket.ChannelId)
		if err != nil {
			// Check if the channel has been deleted
			if restError, ok := err.(request.RestError); ok {
				if restError.StatusCode == 404 {
					if err := dbclient.Client.Tickets.CloseByChannel(*ticket.ChannelId); err != nil {
						ctx.HandleError(err)
						return
					}

					continue
				} else if restError.StatusCode == 403 {
					break
				}
			} else {
				ctx.HandleError(err)
				return
			}
		}

		// Apply overwrites to existing channels
		overwrites := append(ch.PermissionOverwrites, channel.PermissionOverwrite{
			Id:    id,
			Type:  mentionableType.OverwriteType(),
			Allow: permission.BuildPermissions(logic.StandardPermissions[:]...),
			Deny:  0,
		})

		data := rest.ModifyChannelData{
			PermissionOverwrites: overwrites,
			Position:             ch.Position,
		}

		if _, err = ctx.Worker().ModifyChannel(*ticket.ChannelId, data); err != nil {
			ctx.HandleError(err)
			return
		}
	}
}
