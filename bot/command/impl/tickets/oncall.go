package tickets

import (
	"errors"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest/request"
)

type OnCallCommand struct {
}

func (OnCallCommand) Properties() registry.Properties {
	return registry.Properties{
		Name:            "on-call",
		Description:     i18n.HelpAdd,
		Type:            interaction.ApplicationCommandTypeChatInput,
		PermissionLevel: permcache.Support,
		Category:        command.Tickets,
	}
}

func (c OnCallCommand) GetExecutor() interface{} {
	return c.Execute
}

// TODO: Remove on call role from user when removed from team, or they will keep being added to tickets
func (OnCallCommand) Execute(ctx registry.CommandContext) {
	settings, err := dbclient.Client.Settings.Get(ctx.GuildId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if !settings.UseThreads {
		ctx.ReplyPlain("`/on-call` can only be used in thread mode. Visit our docs to find out more") // TODO: Translate
		return
	}

	member, err := ctx.Member()
	if err != nil {
		ctx.HandleError(err)
		return
	}

	// Reflects *new* state
	onCall, err := dbclient.Client.OnCall.Toggle(ctx.GuildId(), ctx.UserId())
	if err != nil {
		ctx.HandleError(err)
		return
	}

	defaultTeam, teamIds, err := logic.GetMemberTeamsWithMember(ctx.GuildId(), ctx.UserId(), member)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	teams, err := dbclient.Client.SupportTeam.GetMulti(ctx.GuildId(), teamIds)
	if err != nil {
		ctx.HandleError(err)
		return
	}

	if onCall { // *new* value
		if defaultTeam {
			if err := assignOnCallRole(ctx, member, settings.OnCallRole, nil, 0); err != nil {
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

		// TODO: Add assigning roles message
		ctx.ReplyPlain("You are now on call")
	} else {
		if defaultTeam && settings.OnCallRole != nil {
			if err := ctx.Worker().RemoveGuildMemberRole(ctx.GuildId(), ctx.UserId(), *settings.OnCallRole); err != nil {
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

		ctx.ReplyPlain("You are no longer on call")
	}
}

// Attempt counter to prevent infinite loop
func assignOnCallRole(ctx registry.CommandContext, member member.Member, roleId *uint64, team *database.SupportTeam, attempt int) error {
	if attempt >= 2 {
		return errors.New("reached retry limit")
	}

	// Create role if it does not exist  yet
	if roleId == nil {
		tmp, err := logic.CreateOnCallRole(ctx, team)
		if err != nil {
			return err
		}

		roleId = &tmp
	}

	if err := ctx.Worker().AddGuildMemberRole(ctx.GuildId(), ctx.UserId(), *roleId); err != nil {
		// If role was deleted, recreate it
		if err, ok := err.(request.RestError); ok && err.StatusCode == 404 && err.ApiError.Message == "Unknown Role" {
			if team == nil {
				if err := dbclient.Client.Settings.SetOnCallRole(ctx.GuildId(), nil); err != nil {
					return err
				}
			} else {
				if err := dbclient.Client.SupportTeam.SetOnCallRole(team.Id, nil); err != nil {
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
