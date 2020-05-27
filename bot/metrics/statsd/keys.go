package statsd

type Key string

const (
	MESSAGES  Key = "messages"
	TICKETS   Key = "tickets"
	COMMANDS  Key = "commands"
	JOINS     Key = "joins"
	LEAVES    Key = "leaves"
	REST      Key = "rest"
	RECONNECT Key = "reconnect"
	IDENTIFY  Key = "identify"
)

func (k Key) String() string {
	return string(k)
}

