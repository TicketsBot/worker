package listeners

import (
	"fmt"
	"github.com/TicketsBot/common/sentry"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/modmail/logic"
	modmailutils "github.com/TicketsBot/worker/bot/modmail/utils"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnDirectOpenMessageReact(worker *worker.Context, e *events.MessageReactionAdd) {
	if e.GuildId != 0 { // DMs only
		return
	}

	if e.UserId == worker.BotId { // ignore our own reactions
		return
	}

	session, err := dbclient.Client.ModmailSession.GetByUser(worker.BotId, e.UserId)
	if err != nil {
		sentry.Error(err)
		return
	}

	if session.UserId != 0 {
		return
	}

	// Determine which emoji was used
	reaction := -1
	for i, emoji := range modmailutils.Emojis {
		if emoji == e.Emoji.Name {
			reaction = i
			break
		}
	}

	// Check a number emoji was used
	if reaction == -1 {
		return
	}

	// Remove reaction
	_ = worker.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, e.Emoji.Name)

	// Create DM channel
	dmChannel, err := worker.CreateDM(e.UserId)
	if err != nil {
		// TODO: Error logging
		return
	}

	// Determine which guild we should open the channel in
	guilds := modmailutils.GetMutualGuilds(worker, e.UserId)

	if reaction-1 >= len(guilds) {
		return
	}

	targetGuild := guilds[reaction-1]

	// Check blacklist
	isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(targetGuild.Id, e.UserId)
	if err != nil {
		sentry.Error(err)
	}

	if isBlacklisted {
		utils.SendEmbed(worker, dmChannel.Id, targetGuild.Id, utils.Red, "Error", translations.MessageBlacklisted, nil, 30, true)
		return
	}

	// Get user object
	user, err := worker.GetUser(e.UserId)
	if err != nil {
		sentry.Error(err)
		return
	}

	utils.SendEmbed(worker, dmChannel.Id, targetGuild.Id, utils.Green, "Modmail", translations.MessageModmailOpened, nil, 0, true, targetGuild.Name)

	// Send guild's welcome message
	welcomeMessage, err := dbclient.Client.WelcomeMessages.Get(targetGuild.Id); if err != nil {
		sentry.Error(err)
		welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
	}

	welcomeMessageId, err := utils.SendEmbedWithResponse(worker, dmChannel.Id, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
	if err != nil {
		utils.SendEmbedRaw(worker, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
		return
	}

	staffChannel, err := logic.OpenModMailTicket(worker, targetGuild, user, welcomeMessageId.Id)
	if err != nil {
		utils.SendEmbedRaw(worker, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
		return
	}

	utils.SendEmbedRaw(worker, staffChannel, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
}

