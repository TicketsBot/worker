package registry

type Flag int

const (
	DMsAllowed Flag = iota
	GuildAllowed
	CanEdit
)

func (f Flag) Int() int {
	return int(f)
}
