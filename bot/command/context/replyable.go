package context

import (
	"fmt"
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/customisation"
	"github.com/TicketsBot/worker/bot/logic"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/TicketsBot/worker/config"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild/emoji"
	"github.com/rxdn/gdl/objects/interaction/component"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest/request"
	"strings"
)

type Replyable struct {
	ctx         registry.CommandContext
	colourCodes map[customisation.Colour]int
}

func NewReplyable(ctx registry.CommandContext) *Replyable {
	var colourCodes map[customisation.Colour]int
	if ctx.PremiumTier() > premium.None {
		var err error
		colourCodes, err = customisation.GetColours(ctx.GuildId())
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			colourCodes = customisation.DefaultColours
		}
	} else {
		colourCodes = customisation.DefaultColours
	}

	return &Replyable{
		ctx:         ctx,
		colourCodes: colourCodes,
	}
}

func (r *Replyable) GetColour(colour customisation.Colour) int {
	return r.colourCodes[colour]
}

func (r *Replyable) buildEmbed(colour customisation.Colour, title, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) *embed.Embed {
	return utils.BuildEmbed(r.ctx, colour, title, content, fields, format...)
}

func (r *Replyable) buildEmbedRaw(colour customisation.Colour, title, content string, fields ...embed.EmbedField) *embed.Embed {
	return utils.BuildEmbedRaw(r.GetColour(colour), title, content, fields, r.ctx.PremiumTier())
}

func (r *Replyable) Reply(colour customisation.Colour, title, content i18n.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPermanent(colour customisation.Colour, title, content i18n.MessageId, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, nil, format...)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbed(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithEmbedAndComponents(embed *embed.Embed, components []component.Component) {
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(embed, components))
}

func (r *Replyable) ReplyWithEmbedPermanent(embed *embed.Embed) {
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFields(colour customisation.Colour, title, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyWithFieldsPermanent(colour customisation.Colour, title, content i18n.MessageId, fields []embed.EmbedField, format ...interface{}) {
	embed := r.buildEmbed(colour, title, content, fields, format...)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRaw(colour customisation.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyRawWithComponents(colour customisation.Colour, title, content string, components ...component.Component) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(command.NewEphemeralEmbedMessageResponseWithComponents(embed, components))
}

func (r *Replyable) ReplyRawPermanent(colour customisation.Colour, title, content string) {
	embed := r.buildEmbedRaw(colour, title, content)
	_, _ = r.ctx.ReplyWith(command.NewEmbedMessageResponse(embed))
}

func (r *Replyable) ReplyPlain(content string) {
	_, _ = r.ctx.ReplyWith(command.NewEphemeralTextMessageResponse(content))
}

func (r *Replyable) ReplyPlainPermanent(content string) {
	_, _ = r.ctx.ReplyWith(command.NewTextMessageResponse(content))
}

func (r *Replyable) HandleError(err error) {
	eventId := sentry.ErrorWithContext(err, r.ctx.ToErrorContext())

	// We should show the invite link if the user is staff (or if we failed to resolve their permission level, show it)
	permLevel, resolveError := r.ctx.UserPermissionLevel()
	showInviteLink := !r.ctx.Worker().IsWhitelabel && (resolveError != nil || permLevel > permcache.Everyone)

	res := r.buildErrorResponse(err, eventId, showInviteLink)
	_, _ = r.ctx.ReplyWith(res)
}

func (r *Replyable) HandleWarning(err error) {
	eventId := sentry.LogWithContext(err, r.ctx.ToErrorContext())

	// We should show the invite link if the user is staff (or if we failed to resolve their permission level, show it)
	permLevel, resolveError := r.ctx.UserPermissionLevel()
	showInviteLink := !r.ctx.Worker().IsWhitelabel && (resolveError != nil || permLevel > permcache.Everyone)

	res := r.buildErrorResponse(err, eventId, showInviteLink)
	_, _ = r.ctx.ReplyWith(res)
}

func (r *Replyable) GetMessage(messageId i18n.MessageId, format ...interface{}) string {
	return i18n.GetMessageFromGuild(r.ctx.GuildId(), messageId, format...)
}

func (r *Replyable) SelectValidEmoji(customEmoji customisation.CustomEmoji, fallback string) *emoji.Emoji {
	if r.ctx.Worker().IsWhitelabel {
		return utils.BuildEmoji(fallback) // TODO: Check whitelabel_guilds table for emojis server
	} else {
		return customEmoji.BuildEmoji()
	}
}

func (r *Replyable) buildErrorResponse(err error, eventId string, includeInviteLink bool) command.MessageResponse {
	var message string
	var imageUrl *string

	if restError, ok := err.(request.RestError); ok {
		if restError.ApiError.Message == "Missing Permissions" { // Override for missing permissions
			interactionCtx, ok := r.ctx.(registry.InteractionContext)
			if ok {
				missingPermissions, err := findMissingPermissions(interactionCtx)
				if err == nil {
					if len(missingPermissions) > 0 {
						message = "I am missing the following permissions required to perform this action:\n"
						for _, perm := range missingPermissions {
							message += fmt.Sprintf("• `%s`\n", perm.String())
						}

						message += "\nPlease assign these permissions to the bot, or alternatively, the `Administrator` permission and try again."
					} else {
						message = formatDiscordError(restError, eventId)
					}
				} else {
					sentry.ErrorWithContext(err, r.ctx.ToErrorContext())
					message = formatDiscordError(restError, eventId)
				}
			} else {
				message = formatDiscordError(restError, eventId)
			}
		} else if restError.ApiError.Message == "CHANNEL_PARENT_INVALID" {
			message = fmt.Sprintf("Invalid channel category. Tell an administrator to visit the [dashboard](https://panel.ticketsbot.net/manage/%d/panels) and assign a valid channel category to this ticket panel.\nError ID: `%s`", r.ctx.GuildId(), eventId)
			imageUrl = utils.Ptr("https://docs.ticketsbot.net/img/multi_panel_category.png")
		} else {
			message = formatDiscordError(restError, eventId)
		}
	} else {
		message = fmt.Sprintf("An error occurred while performing this action.\nError ID: `%s`", eventId)
	}

	embed := r.buildEmbedRaw(customisation.Red, r.GetMessage(i18n.Error), message)
	if imageUrl != nil {
		embed.SetImage(*imageUrl)
	}

	res := command.NewEphemeralEmbedMessageResponse(embed)

	if includeInviteLink {
		res.Components = []component.Component{
			component.BuildActionRow(
				component.BuildButton(component.Button{
					Label: r.GetMessage(i18n.MessageJoinSupportServer),
					Style: component.ButtonStyleLink,
					Emoji: utils.BuildEmoji("❓"),
					Url:   utils.Ptr(strings.ReplaceAll(config.Conf.Bot.SupportServerInvite, "\n", "")),
				}),
			),
		}
	}

	return res
}

func formatDiscordError(restError request.RestError, eventId string) string {
	return fmt.Sprintf("An error occurred while performing this action:\n"+
		"```\n"+
		"%s\n"+
		"```\n"+
		"Error ID: `%s`",
		restError.Error(), eventId)
}

func findMissingPermissions(ctx registry.InteractionContext) ([]permission.Permission, error) {
	if permission.HasPermissionRaw(ctx.InteractionMetadata().AppPermissions, permission.Administrator) {
		return nil, nil
	}

	requiredPermissions := append(
		[]permission.Permission{permission.ManageChannels, permission.ManageRoles},
		logic.StandardPermissions[:]...,
	)

	var missingPermissions []permission.Permission
	for _, requiredPermission := range requiredPermissions {
		if !permission.HasPermissionRaw(ctx.InteractionMetadata().AppPermissions, requiredPermission) {
			missingPermissions = append(missingPermissions, requiredPermission)
		}
	}

	return missingPermissions, nil
}
