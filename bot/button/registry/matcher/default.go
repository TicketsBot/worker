package matcher

type DefaultMatcher struct {
}

func NewDefaultMatcher(prefix string) *DefaultMatcher {
	return &DefaultMatcher{}
}

func (m *DefaultMatcher) Type() Type {
	return TypeDefault
}
