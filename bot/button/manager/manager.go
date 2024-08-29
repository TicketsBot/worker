package manager

import (
	"github.com/TicketsBot/worker/bot/button/handlers"
	"github.com/TicketsBot/worker/bot/button/registry"
	"github.com/TicketsBot/worker/bot/button/registry/matcher"
)

type ComponentInteractionManager struct {
	buttonRegistry registry.ButtonRegistry
	selectRegistry registry.SelectRegistry
	modalRegistry  registry.ModalRegistry

	// button matching engines
	buttonSimpleMatches  map[string]registry.ButtonHandler
	buttonFuncMatches    map[registry.ButtonHandler]matcher.FuncMatchEngine
	buttonDefaultHandler registry.ButtonHandler

	// select menu matching engines
	selectSimpleMatches map[string]registry.SelectHandler
	selectFuncMatches   map[registry.SelectHandler]matcher.FuncMatchEngine

	// modal matching engines
	modalSimpleMatches map[string]registry.ModalHandler
	modalFuncMatches   map[registry.ModalHandler]matcher.FuncMatchEngine
}

func NewButtonManager() *ComponentInteractionManager {
	return &ComponentInteractionManager{
		buttonRegistry: make(registry.ButtonRegistry, 0),
		selectRegistry: make(registry.SelectRegistry, 0),
		modalRegistry:  make(registry.ModalRegistry, 0),

		buttonSimpleMatches:  make(map[string]registry.ButtonHandler),
		buttonFuncMatches:    make(map[registry.ButtonHandler]matcher.FuncMatchEngine),
		buttonDefaultHandler: nil,

		selectSimpleMatches: make(map[string]registry.SelectHandler),
		selectFuncMatches:   make(map[registry.SelectHandler]matcher.FuncMatchEngine),

		modalSimpleMatches: make(map[string]registry.ModalHandler),
		modalFuncMatches:   make(map[registry.ModalHandler]matcher.FuncMatchEngine),
	}
}

func (m *ComponentInteractionManager) GetCommands() []registry.ButtonHandler {
	return m.buttonRegistry
}

func (m *ComponentInteractionManager) RegisterCommands() {
	m.buttonRegistry = append(m.buttonRegistry,
		new(handlers.AddAdminHandler),
		new(handlers.AddSupportHandler),
		new(handlers.CloseHandler),
		new(handlers.CloseWithReasonModalHandler),
		new(handlers.ClaimHandler),
		new(handlers.CloseConfirmHandler),
		new(handlers.CloseRequestAcceptHandler),
		new(handlers.CloseRequestDenyHandler),
		new(handlers.JoinThreadHandler),
		new(handlers.OpenSurveyHandler),
		new(handlers.PanelHandler),
		new(handlers.PremiumCheckAgain),
		new(handlers.PremiumKeyButtonHandler),
		new(handlers.RateHandler),
		new(handlers.RedeemVoteCreditsHandler),
		new(handlers.ViewStaffHandler),
		new(handlers.ViewSurveyHandler),
	)

	m.selectRegistry = append(m.selectRegistry,
		new(handlers.LanguageSelectorHandler),
		new(handlers.MultiPanelHandler),
		new(handlers.PremiumKeyOpenHandler),
	)

	m.modalRegistry = append(m.modalRegistry,
		new(handlers.FormHandler),
		new(handlers.CloseWithReasonSubmitHandler),
		new(handlers.ExitSurveySubmitHandler),
		new(handlers.PremiumKeySubmitHandler),
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

	for _, handler := range m.modalRegistry {
		switch engine := handler.Matcher().(type) {
		case *matcher.SimpleMatcher:
			m.modalSimpleMatches[engine.CustomId] = handler
		case *matcher.FuncMatcher:
			m.modalFuncMatches[handler] = engine.Func
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

func (m *ComponentInteractionManager) MatchModal(customId string) registry.ModalHandler {
	// Try simple match first
	handler, ok := m.modalSimpleMatches[customId]
	if ok {
		return handler
	}

	for handler, f := range m.modalFuncMatches {
		if f(customId) {
			return handler
		}
	}

	return nil
}
