package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker/bot/metrics/prometheus"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"net/http"
	"strconv"
	"strings"
)

var (
	blacklistedHeaders        = []string{"user-agent", "x-real-ip", "cache-control", "content-type", "content-length", "expect", "max-forwards", "pragma", "range", "te", "if-match", "if-none-match", "if-modified-since", "if-unmodified-since", "if-range", "accept", "from", "referer"}
	blacklistedHeaderPrefixes = []string{
		"x-forwarded-",
		"x-proxy-",
		"cf-",
	}

	ErrIntegrationReturnedErrorStatus = errors.New("Integration returned an error status")
)

type formAnswers map[string]*string

type integrationWebhookBody struct {
	GuildId         uint64      `json:"guild_id,string"`
	UserId          uint64      `json:"user_id,string"`
	TicketId        int         `json:"ticket_id"`
	TicketChannelId *uint64     `json:"ticket_channel_id,string"`
	IsNewTicket     bool        `json:"is_new_ticket"`
	FormData        formAnswers `json:"form_data,omitempty"`
}

func Fetch(
	ctx context.Context,
	integration database.CustomIntegration,
	ticket database.Ticket,
	secrets []database.SecretWithValue,
	headers []database.CustomIntegrationHeader,
	placeholders []database.CustomIntegrationPlaceholder, // Only include placeholders that are actually used
	formAnswers formAnswers,
) (map[string]string, error) {
	prometheus.LogIntegrationRequest(integration, ticket.GuildId)

	url := strings.ReplaceAll(integration.WebhookUrl, "%user_id%", strconv.FormatUint(ticket.UserId, 10))
	url = strings.ReplaceAll(url, "%guild_id%", strconv.FormatUint(ticket.GuildId, 10))
	for _, secret := range secrets {
		url = strings.ReplaceAll(url, "%"+secret.Name+"%", secret.Value)
	}

	// Apply headers
	headerMap := make(map[string]string)
	for _, header := range headers {
		if isHeaderBlacklisted(header.Name) {
			continue
		}

		value := header.Value
		value = strings.ReplaceAll(value, "%user_id%", strconv.FormatUint(ticket.UserId, 10))
		value = strings.ReplaceAll(value, "%guild_id%", strconv.FormatUint(ticket.GuildId, 10))
		for _, secret := range secrets {
			value = strings.ReplaceAll(value, "%"+secret.Name+"%", secret.Value)
		}

		headerMap[header.Name] = value
	}

	var body requestBody = nil
	if integration.HttpMethod == http.MethodPost {
		postBody := integrationWebhookBody{
			GuildId:         ticket.GuildId,
			UserId:          ticket.UserId,
			TicketId:        ticket.Id,
			TicketChannelId: ticket.ChannelId,
			IsNewTicket:     true,
		}

		if !integration.Public {
			postBody.FormData = formAnswers
		}

		body = postBody
	}

	res, err := SecureProxy.DoRequest(ctx, integration.HttpMethod, url, headerMap, body)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewBuffer(res))
	decoder.UseNumber()

	var jsonBody map[string]any
	if err := decoder.Decode(&jsonBody); err != nil {
		return nil, err
	}

	return parseBody(jsonBody, placeholders), nil
}

func parseBody(body map[string]any, placeholders []database.CustomIntegrationPlaceholder) map[string]string {
	parsed := make(map[string]string)

outer:
	for _, placeholder := range placeholders {
		parsed[placeholder.Name] = "N/A"

		current := body
		split := strings.Split(placeholder.JsonPath, ".")
		for i, key := range split {
			if i == len(split)-1 {
				value, ok := current[key]
				if ok {
					parsed[placeholder.Name] = fmt.Sprintf("%v", value)
				}
			} else {
				nested, ok := current[key]
				if !ok {
					continue outer
				}

				nestedMap, ok := nested.(map[string]any)
				if current == nil {
					continue outer
				}

				current = nestedMap
			}
		}
	}

	return parsed
}

func isHeaderBlacklisted(name string) bool {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")

	if utils.Contains(blacklistedHeaders, name) {
		return true
	}

	for _, blacklistedPrefix := range blacklistedHeaderPrefixes {
		if strings.HasPrefix(name, blacklistedPrefix) {
			return true
		}
	}

	if name == config.Conf.WebProxy.AuthHeaderName {
		return true
	}

	return false
}
