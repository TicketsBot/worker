package registry

import (
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
	"github.com/TicketsBot/worker/bot/command/context"
)

type ButtonHandler interface {
	Matcher() matcher.Matcher
	Properties() Properties
	Execute(ctx *context.ButtonContext)
}

type Properties struct {
	Flags int
}

func (p *Properties) HasFlag(flag Flag) bool {
	return p.Flags&flag.Int() == flag.Int()
}

func SumFlags(flags ...Flag) (sum int) {
	for _, flag := range flags {
		sum |= flag.Int()
	}

	return sum
}
