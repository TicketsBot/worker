package command

import (
	"context"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/sentry"
	"github.com/TicketsBot/worker/bot/utils"
	"golang.org/x/sync/errgroup"
	"strings"
	"sync"
)

func (ctx *CommandContext) MentionsStaff() bool {
	var lock sync.Mutex
	var mentionsStaff bool

	group, _ := errgroup.WithContext(context.Background())

	for _, user := range ctx.Message.Mentions {
		user.Member.User = user.User

		group.Go(func() error {
			if permission.GetPermissionLevel(utils.ToRetriever(ctx.Worker), user.Member, ctx.GuildId) > permission.Everyone {
				lock.Lock()
				mentionsStaff = true
				lock.Unlock()
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return true
	}

	return mentionsStaff
}

func (ctx *CommandContext) GetMentionedStaff() (userId uint64, found bool) {
	if len(ctx.Mentions) > 0 {
		return ctx.Mentions[0].User.Id, true
	}

	if len(ctx.Args) == 0 {
		return
	}

	// get staff
	supportUsers, err := dbclient.Client.Permissions.GetSupport(ctx.GuildId); if err != nil {
		return
	}

	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(ctx.GuildId); if err != nil {
		return
	}

	query := `SELECT users.user_id FROM users WHERE LOWER("data"->>'Username') LIKE LOWER($1) AND EXISTS(SELECT FROM members WHERE members.guild_id=$2);`
	rows, err := ctx.Worker.Cache.Query(context.Background(), query, strings.Join(ctx.Args, " "), ctx.GuildId)
	defer rows.Close()
	if err != nil {
		return
	}

	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}

		// Check if support rep
		for _, supportUser := range supportUsers {
			if supportUser == id {
				return id, true
			}
		}

		// Check if has support role
		// Get user object
		if member, err := ctx.Worker.GetGuildMember(ctx.GuildId, id); err == nil {
			for _, role := range member.Roles {
				for _, supportRole := range supportRoles {
					if role == supportRole {
						return id, true
					}
				}
			}
		}
	}

	return
}
