package listeners

import "github.com/rxdn/gdl/gateway/payloads/events"

var Listeners = map[events.EventType][]interface{}{
	events.CHANNEL_DELETE:       {OnChannelDelete},
	events.MESSAGE_REACTION_ADD: {OnViewStaffReact},
	events.MESSAGE_CREATE:       {GetCommandListener(), OnMessage, OnSetupProgress},
	events.GUILD_CREATE:         {OnGuildCreate},
	events.GUILD_DELETE:         {OnGuildLeave},
	events.GUILD_MEMBER_REMOVE:  {OnMemberLeave},
	events.GUILD_ROLE_DELETE:    {OnRoleDelete},
}
