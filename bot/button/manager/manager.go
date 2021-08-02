package manager

import (
	"github.com/TicketsBot/worker/bot/button/handlers"
	"github.com/TicketsBot/worker/bot/button/registry"
)

type ButtonManager struct {
	registry registry.Registry
}

func (m *ButtonManager) GetCommands() []registry.ButtonHandler {
	return m.registry
}

func (m *ButtonManager) RegisterCommands() {
	m.registry = append(m.registry, new(handlers.ClaimHandler))
}
