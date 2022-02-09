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

type SelectHandler interface {
	Matcher() matcher.Matcher
	Properties() Properties
	Execute(ctx *context.SelectMenuContext)
}

type ModalHandler interface {
	Matcher() matcher.Matcher
	Properties() Properties
	Execute(ctx *context.ModalContext)
}
