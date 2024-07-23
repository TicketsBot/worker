package logic

import (
	"context"
	"fmt"
	"github.com/TicketsBot/common/integrations/bloxlink"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/database"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/integrations"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// returns msg id
func SendWelcomeMessage(
	ctx registry.CommandContext,
	ticket database.Ticket,
	subject string,
	panel *database.Panel,
	formData map[database.FormInput]string,
	// Only custom integration placeholders for now - prevent making duplicate requests
	additionalPlaceholders map[string]string,
) (uint64, error) {
	settings, err := dbclient.Client.Settings.Get(ticket.GuildId)
	if err != nil {
		return 0, err
	}

	// Build embeds
	welcomeMessageEmbed, err := BuildWelcomeMessageEmbed(ctx, ticket, subject, panel, additionalPlaceholders)
	if err != nil {
		return 0, err
	}

	embeds := utils.Slice(welcomeMessageEmbed)

	// Put form fields in a separate embed
	fields := getFormDataFields(formData)
	if len(fields) > 0 {
		formAnswersEmbed := embed.NewEmbed().
			SetColor(welcomeMessageEmbed.Color)

		for _, field := range fields {
			formAnswersEmbed.AddField(field.Name, field.Value, field.Inline)
		}

		if ctx.PremiumTier() == premium.None {
			formAnswersEmbed.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
		}

		embeds = append(embeds, formAnswersEmbed)
	}

	buttons := []component.Component{
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleClose),
			CustomId: "close",
			Style:    component.ButtonStyleDanger,
			Emoji:    &emoji.Emoji{Name: "🔒"},
		}),
		component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleCloseWithReason),
			CustomId: "close_with_reason",
			Style:    component.ButtonStyleDanger,
			Emoji:    &emoji.Emoji{Name: "🔒"},
		}),
	}

	if !settings.HideClaimButton && !ticket.IsThread {
		buttons = append(buttons, component.BuildButton(component.Button{
			Label:    ctx.GetMessage(i18n.TitleClaim),
			CustomId: "claim",
			Style:    component.ButtonStyleSuccess,
			Emoji:    &emoji.Emoji{Name: "🙋‍♂️"},
		}))
	}

	data := rest.CreateMessageData{
		Embeds: embeds,
		Components: []component.Component{
			component.BuildActionRow(buttons...),
		},
	}

	// Should never happen
	if ticket.ChannelId == nil {
		return 0, fmt.Errorf("channel is nil")
	}

	msg, err := ctx.Worker().CreateMessageComplex(*ticket.ChannelId, data)
	if err != nil {
		return 0, err
	}

	return msg.Id, nil
}

func BuildWelcomeMessageEmbed(
	ctx registry.CommandContext,
	ticket database.Ticket,
	subject string,
	panel *database.Panel,
	// Only custom integration placeholders for now - prevent making duplicate requests
	additionalPlaceholders map[string]string,
) (*embed.Embed, error) {
	if panel == nil || panel.WelcomeMessageEmbed == nil {
		welcomeMessage, err := dbclient.Client.WelcomeMessages.Get(ticket.GuildId)
		if err != nil {
			return nil, err
		}

		if len(welcomeMessage) == 0 {
			welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
		}

		// Replace variables
		welcomeMessage = DoPlaceholderSubstitutions(welcomeMessage, ctx.Worker(), ticket, additionalPlaceholders)

		return utils.BuildEmbedRaw(ctx.GetColour(customisation.Green), subject, welcomeMessage, nil, ctx.PremiumTier()), nil
	} else {
		data, err := dbclient.Client.Embeds.GetEmbed(*panel.WelcomeMessageEmbed)
		if err != nil {
			return nil, err
		}

		fields, err := dbclient.Client.EmbedFields.GetFieldsForEmbed(*panel.WelcomeMessageEmbed)
		if err != nil {
			return nil, err
		}

		e := BuildCustomEmbed(ctx.Worker(), ticket, data, fields, ctx.PremiumTier() == premium.None, additionalPlaceholders)
		return e, nil
	}
}

func DoPlaceholderSubstitutions(
	message string,
	ctx *worker.Context,
	ticket database.Ticket,
	// Only custom integration placeholders for now - prevent making duplicate requests
	additionalPlaceholders map[string]string,
) string {
	var lock sync.Mutex

	// do DB lookups in parallel
	group, _ := errgroup.WithContext(context.Background())
	for placeholder, f := range substitutions {
		placeholder := placeholder
		f := f

		formatted := fmt.Sprintf("%%%s%%", placeholder)

		if strings.Contains(message, formatted) {
			group.Go(func() error {
				replacement := f(ctx, ticket)

				lock.Lock()
				message = strings.Replace(message, formatted, replacement, -1)
				lock.Unlock()

				return nil
			})
		}
	}

	// Group substitutions
	for _, substitutor := range groupSubstitutions {
		substitutor := substitutor

		contains := false
		for _, placeholder := range substitutor.Placeholders {
			formatted := fmt.Sprintf("%%%s%%", placeholder)
			if strings.Contains(message, formatted) {
				contains = true
				break
			}
		}

		if contains {
			group.Go(func() error {
				replacements := substitutor.F(ctx, ticket)
				if replacements == nil {
					replacements = make(map[string]string)
				}

				// Fill any placeholder with N/A that do not have values
				for _, placeholder := range substitutor.Placeholders {
					if _, ok := replacements[placeholder]; !ok {
						replacements[placeholder] = "N/A"
					}
				}

				lock.Lock()
				for placeholder, replacement := range replacements {
					formatted := fmt.Sprintf("%%%s%%", placeholder)
					message = strings.Replace(message, formatted, replacement, -1)
				}
				lock.Unlock()

				return nil
			})
		}
	}

	for placeholder, replacement := range additionalPlaceholders {
		formatted := fmt.Sprintf("%%%s%%", placeholder)
		lock.Lock()
		message = strings.Replace(message, formatted, replacement, -1)
		lock.Unlock()
	}

	if err := group.Wait(); err != nil {
		sentry.Error(err)
	}

	return message
}

func fetchCustomIntegrationPlaceholders(
	ticket database.Ticket,
	formAnswers map[string]*string,
) (map[string]string, error) {
	// Custom integrations
	guildIntegrations, err := dbclient.Client.CustomIntegrationGuilds.GetGuildIntegrations(ticket.GuildId)
	if err != nil {
		return nil, err
	}

	// Fetch integrations
	if len(guildIntegrations) > 0 {
		integrationIds := make([]int, len(guildIntegrations))
		for i, integration := range guildIntegrations {
			integrationIds[i] = integration.Id
		}

		placeholders, err := dbclient.Client.CustomIntegrationPlaceholders.GetAllActivatedInGuild(ticket.GuildId)
		if err != nil {
			return nil, err
		}

		// Determine which integrations we need to fetch
		placeholderMap := make(map[int][]database.CustomIntegrationPlaceholder) // integration_id -> []Placeholder
		for _, placeholder := range placeholders {
			if _, ok := placeholderMap[placeholder.IntegrationId]; !ok {
				placeholderMap[placeholder.IntegrationId] = []database.CustomIntegrationPlaceholder{}
			}

			placeholderMap[placeholder.IntegrationId] = append(placeholderMap[placeholder.IntegrationId], placeholder)
		}

		secrets, err := dbclient.Client.CustomIntegrationSecretValues.GetAll(ticket.GuildId, integrationIds)
		if err != nil {
			return nil, err
		}

		headers, err := dbclient.Client.CustomIntegrationHeaders.GetAll(integrationIds)
		if err != nil {
			return nil, err
		}

		// Replace placeholders
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		group, _ := errgroup.WithContext(ctx)

		var lock sync.Mutex
		m := make(map[string]string) // Merge responses into 1 map

		for _, integration := range guildIntegrations {
			integration := integration
			integrationSecrets := secrets[integration.Id]

			group.Go(func() error {
				response, err := integrations.Fetch(integration, ticket, integrationSecrets, headers[integration.Id], placeholderMap[integration.Id], formAnswers)
				if err != nil {
					return err
				}

				lock.Lock()
				defer lock.Unlock()

				for key, value := range response {
					m[key] = value
				}

				return nil
			})
		}

		if err := group.Wait(); err != nil {
			return nil, err
		}

		return m, nil
	} else {
		return make(map[string]string), nil
	}
}

// TODO: Error handling
type PlaceholderSubstitutionFunc func(*worker.Context, database.Ticket) string

const substitutionTimeout = time.Millisecond * 1500

var substitutions = map[string]PlaceholderSubstitutionFunc{
	"user": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<@%d>", ticket.UserId)
	},
	"ticket_id": func(ctx *worker.Context, ticket database.Ticket) string {
		return strconv.Itoa(ticket.Id)
	},
	"channel": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<#%d>", ticket.ChannelId)
	},
	"username": func(ctx *worker.Context, ticket database.Ticket) string {
		user, _ := ctx.GetUser(ticket.UserId)
		return user.Username
	},
	"server": func(ctx *worker.Context, ticket database.Ticket) string {
		guild, _ := ctx.GetGuild(ticket.GuildId)
		return guild.Name
	},
	"open_tickets": func(_ *worker.Context, ticket database.Ticket) string {
		open, _ := dbclient.Client.Tickets.GetGuildOpenTickets(ticket.GuildId)
		return strconv.Itoa(len(open))
	},
	"total_tickets": func(_ *worker.Context, ticket database.Ticket) string {
		ctx, cancel := context.WithTimeout(context.Background(), substitutionTimeout)
		defer cancel()

		count, _ := dbclient.Analytics.GetTotalTicketCount(ctx, ticket.GuildId)
		return strconv.FormatUint(count, 10)
	},
	"user_open_tickets": func(ctx *worker.Context, ticket database.Ticket) string {
		count, _ := dbclient.Client.Tickets.GetOpenCountByUser(ticket.GuildId, ticket.UserId)
		return strconv.Itoa(count)
	},
	"user_total_tickets": func(ctx *worker.Context, ticket database.Ticket) string {
		tickets, _ := dbclient.Client.Tickets.GetTotalCountByUser(ticket.GuildId, ticket.UserId)
		return strconv.Itoa(tickets)
	},
	"ticket_limit": func(ctx *worker.Context, ticket database.Ticket) string {
		limit, _ := dbclient.Client.TicketLimit.Get(ticket.GuildId)
		return strconv.Itoa(int(limit))
	},
	"rating_count": func(_ *worker.Context, ticket database.Ticket) string {
		ctx, cancel := context.WithTimeout(context.Background(), substitutionTimeout)
		defer cancel()

		ratingCount, _ := dbclient.Analytics.GetFeedbackCountGuild(ctx, ticket.GuildId)
		return strconv.FormatUint(ratingCount, 10)
	},
	"average_rating": func(_ *worker.Context, ticket database.Ticket) string {
		ctx, cancel := context.WithTimeout(context.Background(), substitutionTimeout)
		defer cancel()

		average, _ := dbclient.Analytics.GetAverageFeedbackRatingGuild(ctx, ticket.GuildId)
		return fmt.Sprintf("%.1f", average)
	},
	"time": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:t>", time.Now().Unix())
	},
	"date": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:d>", time.Now().Unix())
	},
	"datetime": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:f>", time.Now().Unix())
	},
	"first_response_time_weekly": func(ctx *worker.Context, ticket database.Ticket) string {
		if !ctx.IsWhitelabel { // If whitelabel, the bot must be premium, so we don't need to do extra checks
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, ctx.Token, ctx.RateLimiter)
			if err != nil {
				sentry.Error(err)
				return ""
			}

			if premiumTier == premium.None {
				return ""
			}
		}

		context, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
		defer cancel()

		data, err := dbclient.Analytics.GetFirstResponseTimeStats(context, ticket.GuildId)
		if err != nil {
			sentry.Error(err)
			return ""
		}

		return utils.FormatNullableTime(data.Weekly)
	},
	"first_response_time_monthly": func(ctx *worker.Context, ticket database.Ticket) string {
		if !ctx.IsWhitelabel { // If whitelabel, the bot must be premium, so we don't need to do extra checks
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, ctx.Token, ctx.RateLimiter)
			if err != nil {
				sentry.Error(err)
				return ""
			}

			if premiumTier == premium.None {
				return ""
			}
		}

		context, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
		defer cancel()

		data, err := dbclient.Analytics.GetFirstResponseTimeStats(context, ticket.GuildId)
		if err != nil {
			sentry.Error(err)
			return ""
		}

		return utils.FormatNullableTime(data.Monthly)
	},
	"first_response_time_all_time": func(ctx *worker.Context, ticket database.Ticket) string {
		if !ctx.IsWhitelabel { // If whitelabel, the bot must be premium, so we don't need to do extra checks
			premiumTier, err := utils.PremiumClient.GetTierByGuildId(ticket.GuildId, true, ctx.Token, ctx.RateLimiter)
			if err != nil {
				sentry.Error(err)
				return ""
			}

			if premiumTier == premium.None {
				return ""
			}
		}

		context, cancel := context.WithTimeout(context.Background(), time.Millisecond*1500)
		defer cancel()

		data, err := dbclient.Analytics.GetFirstResponseTimeStats(context, ticket.GuildId)
		if err != nil {
			sentry.Error(err)
			return ""
		}

		return utils.FormatNullableTime(data.AllTime)
	},
	"discord_account_creation_date": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:d>", utils.SnowflakeToTime(ticket.UserId).Unix())
	},
	"discord_account_age": func(ctx *worker.Context, ticket database.Ticket) string {
		return fmt.Sprintf("<t:%d:R>", utils.SnowflakeToTime(ticket.UserId).Unix())
	},
}

type GroupSubstitutionFunc func(*worker.Context, database.Ticket) map[string]string

type GroupSubstitutor struct {
	Placeholders []string
	F            GroupSubstitutionFunc
}

func NewGroupSubstitutor(placeholders []string, f GroupSubstitutionFunc) GroupSubstitutor {
	return GroupSubstitutor{
		Placeholders: placeholders,
		F:            f,
	}
}

var groupSubstitutions = []GroupSubstitutor{
	NewGroupSubstitutor([]string{"roblox_username", "roblox_id", "roblox_display_name", "roblox_profile_url", "roblox_account_age", "roblox_account_created"},
		func(ctx *worker.Context, ticket database.Ticket) map[string]string {
			user, err := integrations.Bloxlink.GetRobloxUser(ticket.UserId)
			if err != nil {
				if err == bloxlink.ErrUserNotFound {
					return nil
				} else {
					sentry.Error(err)
					return nil
				}
			}

			return map[string]string{
				"roblox_username":        user.Name,
				"roblox_id":              strconv.Itoa(user.Id),
				"roblox_display_name":    user.DisplayName,
				"roblox_profile_url":     fmt.Sprintf("https://www.roblox.com/users/%d/profile", user.Id),
				"roblox_account_age":     fmt.Sprintf("<t:%d:R>", user.Created.Unix()),
				"roblox_account_created": fmt.Sprintf("<t:%d:D>", user.Created.Unix()),
			}
		},
	),
}

func formAnswersToMap(formData map[database.FormInput]string) map[string]*string {
	// Get form inputs in the same order they are presented on the dashboard
	i := 0
	inputs := make([]database.FormInput, len(formData))
	for input := range formData {
		inputs[i] = input
		i++
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Position < inputs[j].Position
	})

	answers := make(map[string]*string)
	for _, input := range inputs {
		answer, ok := formData[input]
		if ok {
			answers[input.Label] = &answer
		} else {
			answers[input.Label] = nil
		}
	}

	return answers
}

func getFormDataFields(formData map[database.FormInput]string) []embed.EmbedField {
	// Get form inputs in the same order they are presented on the dashboard
	i := 0
	inputs := make([]database.FormInput, len(formData))
	for input := range formData {
		inputs[i] = input
		i++
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Position < inputs[j].Position
	})

	var fields []embed.EmbedField // Can't use len(formData), as form may have changed since modal was opened
	for _, input := range inputs {
		answer, ok := formData[input]
		if answer == "" {
			answer = "N/A" // TODO: What should we use here?
		}

		if ok {
			fields = append(fields, embed.EmbedField{
				Name:   input.Label,
				Value:  answer,
				Inline: false,
			})
		}
	}

	return fields
}

func BuildCustomEmbed(
	ctx *worker.Context,
	ticket database.Ticket,
	customEmbed database.CustomEmbed,
	fields []database.EmbedField,
	branding bool,
	// Only custom integration placeholders for now - prevent making duplicate requests
	additionalPlaceholders map[string]string,
) *embed.Embed {
	description := utils.ValueOrZero(customEmbed.Description)
	if ticket.Id != 0 {
		description = DoPlaceholderSubstitutions(description, ctx, ticket, additionalPlaceholders)
	}

	e := &embed.Embed{
		Title:       utils.ValueOrZero(customEmbed.Title),
		Description: description,
		Url:         utils.ValueOrZero(customEmbed.Url),
		Timestamp:   customEmbed.Timestamp,
		Color:       int(customEmbed.Colour),
	}

	if branding {
		e.SetFooter("Powered by ticketsbot.net", "https://ticketsbot.net/assets/img/logo.png")
	} else if customEmbed.FooterText != nil {
		e.SetFooter(*customEmbed.FooterText, utils.ValueOrZero(customEmbed.FooterIconUrl))
	}

	if customEmbed.ImageUrl != nil {
		e.SetImage(*customEmbed.ImageUrl)
	}

	if customEmbed.ThumbnailUrl != nil {
		e.SetThumbnail(*customEmbed.ThumbnailUrl)
	}

	if customEmbed.AuthorName != nil {
    		authorName := DoPlaceholderSubstitutions(*customEmbed.AuthorName, ctx, ticket, additionalPlaceholders)

    		authorUrl := ValueOrZero(customEmbed.AuthorUrl)
    		if authorUrl != "" {
        		authorUrl = DoPlaceholderSubstitutions(authorUrl, ctx, ticket, additionalPlaceholders)
    		}

    		authorIconUrl := ValueOrZero(customEmbed.AuthorIconUrl)
    		if authorIconUrl != "" {
        		authorIconUrl = DoPlaceholderSubstitutions(authorIconUrl, ctx, ticket, additionalPlaceholders)
    		}

    		e.SetAuthor(authorName, authorUrl, authorIconUrl)
	}



	for _, field := range fields {
		value := field.Value
		if ticket.Id != 0 {
			value = DoPlaceholderSubstitutions(value, ctx, ticket, additionalPlaceholders)
		}

		e.AddField(field.Name, value, field.Inline)
	}

	return e
}
