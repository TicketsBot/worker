package matcher

type FuncMatchEngine func(customId string) bool

type FuncMatcher struct {
	Func FuncMatchEngine
}

func NewFuncMatcher(engine FuncMatchEngine) *FuncMatcher {
	return &FuncMatcher{
		Func: engine,
	}
}

func (m *FuncMatcher) Type() Type {
	return TypeFunc
}
