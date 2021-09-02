package matcher

type Matcher interface {
	Type() Type
}

type Type uint8

const (
	TypeSimple Type = iota
	TypeFunc
	TypeDefault
)


