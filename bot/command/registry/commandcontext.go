package registry

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/constants"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type CommandContext interface {
	Worker() *worker.Context

	GuildId() uint64
	ChannelId() uint64
	UserId() uint64

	UserPermissionLevel() (permcache.PermissionLevel, error)
	PremiumTier() premium.PremiumTier
	IsInteraction() bool
	ToErrorContext() errorcontext.WorkerErrorContext

	Reply(colour constants.Colour, title, content i18n.MessageId, format ...interface{})
	ReplyWith(response command.MessageResponse) (message.Message, error)
	ReplyWithEmbed(embed *embed.Embed)
	ReplyWithEmbedPermanent(embed *embed.Embed)
	ReplyPermanent(colour constants.Colour, title, content i18n.MessageId, format ...interface{})
	ReplyWithFields(colour constants.Colour, title, content i18n.MessageId, fields []embed.EmbedField, format ...interface{})
	ReplyWithFieldsPermanent(colour constants.Colour, title, content i18n.MessageId, fields []embed.EmbedField, format ...interface{})

	ReplyRaw(colour constants.Colour, title, content string)
	ReplyRawPermanent(colour constants.Colour, title, content string)

	ReplyPlain(content string)
	ReplyPlainPermanent(content string)

	// No functionality on interactions, check / cross reaction on messages
	Accept()
	Reject()

	HandleError(err error)
	HandleWarning(err error)

	GetMessage(messageId i18n.MessageId, format ...interface{}) string

	// Utility functions
	Guild() (guild.Guild, error)
	Member() (member.Member, error)
	User() (user.User, error)
}
