package command

type Command interface {
	Execute(ctx CommandContext)
	Properties() Properties
}
