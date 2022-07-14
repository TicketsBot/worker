package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

var (
	IntegrationRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tickets",
		Subsystem: "worker",
		Name:      "integration_requests",
	}, []string{"integration_id", "guild_id"})

	TicketsCreated = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tickets",
		Subsystem: "worker",
		Name:      "tickets_created",
	}, []string{"guild_id"})

	Commands = newCounterVec("commands", []string{"guild_id", "command"})
)

func newCounterVec(name string, labels []string) *prometheus.CounterVec {
	return promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tickets",
		Subsystem: "worker",
		Name:      name,
	}, labels)
}

func LogIntegrationRequest(integrationId int, guildId uint64) {
	IntegrationRequests.WithLabelValues(strconv.Itoa(integrationId), strconv.FormatUint(guildId, 10)).Inc()
}

func LogTicketCreated(guildId uint64) {
	TicketsCreated.WithLabelValues(strconv.FormatUint(guildId, 10)).Inc()
}

func LogCommand(guildId uint64, command string) {
	Commands.WithLabelValues(strconv.FormatUint(guildId, 10), command).Inc()
}
