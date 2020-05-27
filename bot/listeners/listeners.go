package listeners

var Listeners = []interface{}{
	OnChannelDelete,
	OnCloseConfirm,
	OnCloseReact,
	OnCommand,
	OnFirstResponse,
	OnGuildCreate,
	OnGuildLeave,
	OnMemberLeave,
	OnMessage,
	OnPanelReact,
	OnSetupProgress,
}
