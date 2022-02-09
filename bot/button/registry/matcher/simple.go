package matcher

type SimpleMatcher struct {
	CustomId string
}

func NewSimpleMatcher(customId string) *SimpleMatcher {
	return &SimpleMatcher{
		CustomId: customId,
	}
}

func (m *SimpleMatcher) Type() Type {
	return TypeSimple
}
