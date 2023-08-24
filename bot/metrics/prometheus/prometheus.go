package prometheus

import (
	"github.com/TicketsBot/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

var (
	IntegrationRequests = newCounterVec("integration_requests", "integration_id", "integration_name", "guild_id")
	TicketsCreated      = newCounterVec("tickets_created", "guild_id")

	Commands = newCounterVec("commands", "guild_id", "command")
)

func newCounterVec(name string, labels ...string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tickets",
		Subsystem: "worker",
		Name:      name,
	}, labels)
}

func LogIntegrationRequest(integration database.CustomIntegration, guildId uint64) {
	IntegrationRequests.WithLabelValues(
		strconv.Itoa(integration.Id),
		integration.Name,
		strconv.FormatUint(guildId, 10),
	).Inc()
}

func LogTicketCreated(guildId uint64) {
	TicketsCreated.WithLabelValues(strconv.FormatUint(guildId, 10)).Inc()
}

func LogCommand(guildId uint64, command string) {
	Commands.WithLabelValues(strconv.FormatUint(guildId, 10), command).Inc()
}
