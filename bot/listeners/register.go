package listeners

func init() {
	ChannelDeleteListeners = append(ChannelDeleteListeners, OnChannelDelete)
	GuildCreateListeners = append(GuildCreateListeners, OnGuildCreate)
	GuildDeleteListeners = append(GuildDeleteListeners, OnGuildLeave)
	GuildMemberRemoveListeners = append(GuildMemberRemoveListeners, OnMemberLeave)
	GuildMemberUpdateListeners = append(GuildMemberUpdateListeners, OnMemberUpdate)
	MessageCreateListeners = append(MessageCreateListeners, OnMessage)
	GuildRoleDeleteListeners = append(GuildRoleDeleteListeners, OnRoleDelete)
	ThreadMembersUpdateListeners = append(ThreadMembersUpdateListeners, OnThreadMembersUpdate)
	ThreadUpdateListeners = append(ThreadUpdateListeners, OnThreadUpdate)
}
