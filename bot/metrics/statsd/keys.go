package statsd

type Key string

const (
	KeyMessages      Key = "messages"
	KeyTickets       Key = "tickets"
	KeyCommands      Key = "commands"
	KeyJoins         Key = "joins"
	KeyLeaves        Key = "leaves"
	KeyRest          Key = "rest"
	KeySlashCommands Key = "slash_commands"
	KeyEvents        Key = "events"
	AutoClose        Key = "autoclose"
	KeyDirectMessage Key = "direct_message"
	KeyOpenCommand   Key = "open_command"
)

func (k Key) String() string {
	return string(k)
}

func AllKeys() []Key {
	return []Key{
		KeyMessages,
		KeyTickets,
		KeyCommands,
		KeyJoins,
		KeyLeaves,
		KeyRest,
		KeySlashCommands,
		KeyEvents,
		AutoClose,
		KeyDirectMessage,
		KeyOpenCommand,
	}
}
