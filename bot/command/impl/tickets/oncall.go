package tickets

import (
	"errors"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest/request"
	"time"
)

type OnCallCommand struct {
}

func (OnCallCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:             "on-call",
		Description:      i18n.HelpOnCall,
		Type:             interaction.ApplicationCommandTypeChatInput,
		PermissionLevel:  permcache.Support,
		Category:         command.Tickets,
		DefaultEphemeral: true,
		Timeout:          time.Second * 8,
	}
}

func (c OnCallCommand) GetExecutor() interface{} {
	return c.Execute
}

func (OnCallCommand) Execute(ctx registry.CommandContext) {
	settings, err := ctx.Settings()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !settings.UseThreads {
		ctx.Reply(customisation.Red, i18n.Error, i18n.MessageOnCallChannelMode)
		return
	}

	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Reflects *new* state
	onCall, err := dbclient.Client.OnCall.Toggle(ctx, ctx.GuildId(), ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	defaultTeam, teamIds, err := logic.GetMemberTeamsWithMember(ctx, ctx.GuildId(), ctx.UserId(), member)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	teams, err := dbclient.Client.SupportTeam.GetMulti(ctx, ctx.GuildId(), teamIds)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	metadata, err := dbclient.Client.GuildMetadata.Get(ctx, ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if onCall { // *new* value
		if defaultTeam {
			if err := assignOnCallRole(ctx, member, metadata.OnCallRole, nil, 0); err != nil {
				ctx.HandleError(err)
				return
			}
		}

		for i, teamId := range teamIds {
			if i >= 5 { // Don't get caught up adding roles forever
				break
			}

			team, ok := teams[teamId]
			if !ok {
				continue
			}

			if err := assignOnCallRole(ctx, member, team.OnCallRole, &team, 0); err != nil {
				ctx.HandleError(err)
				return
			}
		}

		// TODO: Add assigning roles progress message
		ctx.Reply(customisation.Green, i18n.Success, i18n.MessageOnCallSuccess)
	} else {
		if defaultTeam && metadata.OnCallRole != nil {
			if err := ctx.Worker().RemoveGuildMemberRole(ctx.GuildId(), ctx.UserId(), *metadata.OnCallRole); err != nil {
				ctx.HandleError(err)
				return
			}
		}

		for i, teamId := range teamIds {
			if i >= 5 { // Don't get caught up adding roles forever
				break
			}

			team, ok := teams[teamId]
			if !ok {
				continue
			}

			if team.OnCallRole == nil {
				continue
			}

			if err := ctx.Worker().RemoveGuildMemberRole(ctx.GuildId(), ctx.UserId(), *team.OnCallRole); err != nil {
				ctx.HandleError(err)
				return
			}
		}

		ctx.Reply(customisation.Green, i18n.Success, i18n.MessageOnCallRemoveSuccess)
	}
}

// Attempt counter to prevent infinite loop
func assignOnCallRole(ctx registry.CommandContext, member member.Member, roleId *uint64, team *database.SupportTeam, attempt int) error {
	if attempt >= 2 {
		return errors.New("reached retry limit")
	}

	// Create role if it does not exist  yet
	if roleId == nil {
		tmp, err := logic.CreateOnCallRole(ctx, ctx, team)
		if err != nil {
			return err
		}

		roleId = &tmp
	}

	if err := ctx.Worker().AddGuildMemberRole(ctx.GuildId(), ctx.UserId(), *roleId); err != nil {
		// If role was deleted, recreate it
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 && err.ApiError.Message == "Unknown Role" {
			if team == nil {
				if err := dbclient.Client.GuildMetadata.SetOnCallRole(ctx, ctx.GuildId(), nil); err != nil {
					return err
				}
			} else {
				if err := dbclient.Client.SupportTeam.SetOnCallRole(ctx, team.Id, nil); err != nil {
					return err
				}
			}

			return assignOnCallRole(ctx, member, nil, team, attempt+1)
		} else {
			return err
		}
	}

	return nil
}
