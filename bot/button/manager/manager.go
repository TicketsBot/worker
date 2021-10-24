package manager

import (
	"github.com/TicketsBot/worker/bot/button/handlers"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
)

type ComponentInteractionManager struct {
	buttonRegistry registry.ButtonRegistry
	selectRegistry registry.SelectRegistry

	// button matching engines
	buttonSimpleMatches  map[string]registry.ButtonHandler
	buttonFuncMatches    map[registry.ButtonHandler]matcher.FuncMatchEngine
	buttonDefaultHandler registry.ButtonHandler

	// select menu matching engines
	selectSimpleMatches map[string]registry.SelectHandler
	selectFuncMatches   map[registry.SelectHandler]matcher.FuncMatchEngine
}

func NewButtonManager() *ComponentInteractionManager {
	return &ComponentInteractionManager{
		buttonRegistry: make(registry.ButtonRegistry, 0),
		selectRegistry: make(registry.SelectRegistry, 0),

		buttonSimpleMatches:  make(map[string]registry.ButtonHandler),
		buttonFuncMatches:    make(map[registry.ButtonHandler]matcher.FuncMatchEngine),
		buttonDefaultHandler: nil,

		selectSimpleMatches: make(map[string]registry.SelectHandler),
		selectFuncMatches:   make(map[registry.SelectHandler]matcher.FuncMatchEngine),
	}
}

func (m *ComponentInteractionManager) GetCommands() []registry.ButtonHandler {
	return m.buttonRegistry
}

func (m *ComponentInteractionManager) RegisterCommands() {
	m.buttonRegistry = append(m.buttonRegistry,
		new(handlers.CloseHandler),
		new(handlers.ClaimHandler),
		new(handlers.CloseConfirmHandler),
		new(handlers.CloseRequestAcceptHandler),
		new(handlers.CloseRequestDenyHandler),
		new(handlers.PanelHandler),
		new(handlers.RateHandler),
		new(handlers.ViewStaffHandler),
	)

	m.selectRegistry = append(m.selectRegistry,
		new(handlers.MultiPanelHandler),
	)

	for _, handler := range m.buttonRegistry {
		switch engine := handler.Matcher().(type) {
		case *matcher.SimpleMatcher:
			m.buttonSimpleMatches[engine.CustomId] = handler
		case *matcher.FuncMatcher:
			m.buttonFuncMatches[handler] = engine.Func
		case *matcher.DefaultMatcher:
			m.buttonDefaultHandler = handler
		}
	}

	for _, handler := range m.selectRegistry {
		switch engine := handler.Matcher().(type) {
		case *matcher.SimpleMatcher:
			m.selectSimpleMatches[engine.CustomId] = handler
		case *matcher.FuncMatcher:
			m.selectFuncMatches[handler] = engine.Func
		case *matcher.DefaultMatcher:
			panic("default matcher not allowed for select menu")
		}
	}
}

func (m *ComponentInteractionManager) MatchButton(customId string) registry.ButtonHandler {
	// Try simple match first
	handler, ok := m.buttonSimpleMatches[customId]
	if ok {
		return handler
	}

	for handler, f := range m.buttonFuncMatches {
		if f(customId) {
			return handler
		}
	}

	// Ok to return nil
	return m.buttonDefaultHandler
}

func (m *ComponentInteractionManager) MatchSelect(customId string) registry.SelectHandler {
	// Try simple match first
	handler, ok := m.selectSimpleMatches[customId]
	if ok {
		return handler
	}

	for handler, f := range m.selectFuncMatches {
		if f(customId) {
			return handler
		}
	}

	return nil
}
