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

	OnMessageTicketLookup = newCounterVec("on_message_ticket_lookup_count", "is_ticket", "cache_hit")

	ActiveHttpRequests  = newGauge("active_http_requests")
	HttpRequestDuration = newHistogram("http_request_duration")

	InboundRequests    = newCounterVec("inbound_requests", "route")
	ActiveInteractions = newGauge("active_interactions")
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

func newGauge(name string) prometheus.Gauge {
	return promauto.NewGauge(prometheus.GaugeOpts{
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

func LogOnMessageTicketLookup(isTicket, cacheHit bool) {
	OnMessageTicketLookup.WithLabelValues(strconv.FormatBool(isTicket), strconv.FormatBool(cacheHit)).Inc()
}
