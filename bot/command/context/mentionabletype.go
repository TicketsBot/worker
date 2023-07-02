package context

import (
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/rxdn/gdl/objects"
	"github.com/rxdn/gdl/objects/channel"
)

type MentionableType uint8

const (
	MentionableTypeUser MentionableType = iota
	MentionableTypeRole
)

func (m MentionableType) OverwriteType() channel.PermissionOverwriteType {
	switch m {
	case MentionableTypeUser:
		return channel.PermissionTypeMember
	case MentionableTypeRole:
		return channel.PermissionTypeRole
	default:
		return -1
	}
}

// DetermineMentionableType TODO: Move this function to be a method on the CommandContext interface
// DetermineMentionableType (type, ok)
func DetermineMentionableType(ctx registry.CommandContext, id uint64) (MentionableType, bool) {
	interactionCtx, ok := ctx.(*SlashCommandContext)
	if ok {
		resolved := interactionCtx.Interaction.Data.Resolved
		if _, isUser := resolved.Users[objects.Snowflake(id)]; isUser {
			return MentionableTypeUser, true
		} else if _, isRole := resolved.Roles[objects.Snowflake(id)]; isRole {
			return MentionableTypeRole, true
		} else {
			return 0, false
		}
	} else {
		return 0, false
	}
}
