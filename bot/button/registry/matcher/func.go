package matcher

type FuncMatchEngine func(customId string) bool

type FuncMatcher struct {
	Func FuncMatchEngine
}

func (m *FuncMatcher) Type() Type {
	return TypeFunc
}
