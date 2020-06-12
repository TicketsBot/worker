package listeners

import (
	"fmt"
	"github.com/TicketsBot/common/eventforwarding"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/modmail/logic"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"strings"
)

func OnModMailChannelMessage(worker *worker.Context, e *events.MessageCreate, extra eventforwarding.Extra) {
	if e.Author.Id == worker.BotId {
		return
	}

	if e.GuildId == 0 { // Guilds only
		return
	}

	errorContext := errorcontext.WorkerErrorContext{
		Guild:   e.GuildId,
		Channel: e.ChannelId,
		Shard:   worker.ShardId,
		User:    e.Author.Id,
	}

	session, err := dbclient.Client.ModmailSession.GetByChannel(worker.BotId, e.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if session.UserId == 0 {
		return
	}

	// TODO: Make this less hacky
	// check close
	if isClose, args := isClose(e); isClose {
		// get permission level
		permLevel := permission.GetPermissionLevel(utils.ToRetriever(worker), e.Member, e.GuildId)

		logic.HandleClose(session, command.CommandContext{
			Worker:              worker,
			Message:             e.Message,
			Root:                "close",
			Args:                args,
			PremiumTier:         utils.PremiumClient.GetTierByGuildId(e.GuildId, true, worker.Token, worker.RateLimiter),
			ShouldReact:         true,
			IsFromPanel:         false,
			UserPermissionLevel: permLevel,
		})
		return
	}

	// Make sure we don't mirror the user's message back to them
	var username string
	if user, found := worker.Cache.GetUser(session.UserId); found {
		username = user.Username
	}

	// TODO: Make this less hacky
	if e.Author.Username == username && e.WebhookId != 0 {
		return
	}

	// Create DM channel
	privateMessageChannel, err := worker.CreateDM(session.UserId)
	if err != nil { // User probably has DMs disabled
		sentry.LogWithContext(err, errorContext)
		return
	}

	message := fmt.Sprintf("**%s**: %s", e.Author.Username, e.Message.Content)
	if _, err := worker.CreateMessage(privateMessageChannel.Id, message); err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// forward attachments
	// don't re-upload attachments incase user has uploaded TOS breaking attachment
	if len(e.Message.Attachments) > 0 {
		var content string
		if len(e.Message.Attachments) == 1 {
			content = fmt.Sprintf("%s attached a file:", e.Author.Mention())
		} else {
			content = fmt.Sprintf("%s attached files:", e.Author.Mention())
		}

		for _, attachment := range e.Message.Attachments {
			content += fmt.Sprintf("\n▶️ %s", attachment.ProxyUrl)
		}

		if _, err := worker.CreateMessage(privateMessageChannel.Id, content); err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}
	}
}

// isClose, args
func isClose(e *events.MessageCreate) (bool, []string) {
	customPrefix, err := dbclient.Client.Prefix.Get(e.GuildId)
	if err != nil {
		sentry.Error(err)
	}

	var usedPrefix string

	if strings.HasPrefix(strings.ToLower(e.Content), utils.DEFAULT_PREFIX) {
		usedPrefix = utils.DEFAULT_PREFIX
	} else if customPrefix != "" && strings.HasPrefix(e.Content, customPrefix) {
		usedPrefix = customPrefix
	} else { // Not a command
		return false, nil
	}

	split := strings.Split(e.Content, " ")
	root := split[0][len(usedPrefix):]

	if strings.ToLower(root) != "close" {
		return false, nil
	}

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	return true, args
}
