package manager

import (
	"github.com/TicketsBot/worker/bot/button/handlers"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
)

type ButtonManager struct {
	registry registry.Registry

	// matching engines
	simpleMatches  map[string]registry.ButtonHandler
	funcMatches    map[registry.ButtonHandler]matcher.FuncMatchEngine
	defaultHandler registry.ButtonHandler
}

func NewButtonManager() *ButtonManager {
	return &ButtonManager{
		registry:       make(registry.Registry, 0),
		simpleMatches:  make(map[string]registry.ButtonHandler),
		funcMatches:    make(map[registry.ButtonHandler]matcher.FuncMatchEngine),
		defaultHandler: nil,
	}
}

func (m *ButtonManager) GetCommands() []registry.ButtonHandler {
	return m.registry
}

func (m *ButtonManager) RegisterCommands() {
	m.registry = append(m.registry,
		new(handlers.CloseHandler),
		new(handlers.ClaimHandler),
		new(handlers.CloseConfirmHandler),
		new(handlers.CloseRequestAcceptHandler),
		new(handlers.CloseRequestDenyHandler),
		new(handlers.PanelHandler),
		new(handlers.RateHandler),
		new(handlers.ViewStaffHandler),
	)

	for _, handler := range m.registry {
		switch engine := handler.Matcher().(type) {
		case *matcher.SimpleMatcher:
			m.simpleMatches[engine.CustomId] = handler
		case *matcher.FuncMatcher:
			m.funcMatches[handler] = engine.Func
		case *matcher.DefaultMatcher:
			m.defaultHandler = handler
		}
	}
}

func (m *ButtonManager) Match(customId string) registry.ButtonHandler {
	// Try simple match first
	handler, ok := m.simpleMatches[customId]
	if ok {
		return handler
	}

	for handler, f := range m.funcMatches {
		if f(customId) {
			return handler
		}
	}

	// Ok to return nil
	return m.defaultHandler
}
