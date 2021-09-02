package matcher

type SimpleMatcher struct {
	CustomId string
}

func (m *SimpleMatcher) Type() Type {
	return TypeSimple
}
