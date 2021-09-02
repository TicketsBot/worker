package matcher

type DefaultMatcher struct {
}

func (m *DefaultMatcher) Type() Type {
	return TypeDefault
}
