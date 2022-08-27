package listeners

import "github.com/rxdn/gdl/gateway/payloads/events"

var Listeners = map[events.EventType][]interface{}{
	events.CHANNEL_DELETE:        {OnChannelDelete},
	events.MESSAGE_CREATE:        {GetCommandListener(), OnMessage},
	events.GUILD_CREATE:          {OnGuildCreate},
	events.GUILD_DELETE:          {OnGuildLeave},
	events.GUILD_MEMBER_UPDATE:   {OnMemberUpdate},
	events.GUILD_MEMBER_REMOVE:   {OnMemberLeave},
	events.GUILD_ROLE_DELETE:     {OnRoleDelete},
	events.THREAD_UPDATE:         {OnThreadUpdate},
	events.THREAD_MEMBERS_UPDATE: {OnThreadMembersUpdate},
}
