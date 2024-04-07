package prometheus

import (
	"github.com/TicketsBot/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

const (
	Namespace = "tickets"
	Subsystem = "worker"
)

var (
	IntegrationRequests = newCounterVec("integration_requests", "integration_id", "integration_name", "guild_id")
	TicketsCreated      = newCounterVec("tickets_created", "guild_id")

	Commands = newCounterVec("commands", "guild_id", "command")

	InteractionTimeToDefer   = newHistogram("interaction_time_to_defer")
	InteractionTimeToReceive = newHistogram("interaction_time_to_receive")
)

func newCounterVec(name string, labels ...string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      name,
	}, labels)
}

func newHistogram(name string) prometheus.Histogram {
	return promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: Subsystem,
		Name:      name,
	})
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
