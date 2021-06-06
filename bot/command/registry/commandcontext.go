package registry

import (
	permcache "github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	translations "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker"
	"github.com/TicketsBot/worker/bot/errorcontext"
	"github.com/TicketsBot/worker/bot/utils"
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

	Reply(colour utils.Colour, title string, content translations.MessageId, format ...interface{})
	ReplyWithEmbed(embed *embed.Embed)
	ReplyWithEmbedPermanent(embed *embed.Embed) (message.Message, error)
	ReplyPermanent(colour utils.Colour, title string, content translations.MessageId, format ...interface{})
	ReplyWithFields(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{})
	ReplyWithFieldsPermanent(colour utils.Colour, title string, content translations.MessageId, fields []embed.EmbedField, format ...interface{})

	ReplyRaw(colour utils.Colour, title, content string)
	ReplyRawPermanent(colour utils.Colour, title, content string)

	ReplyPlain(content string)
	ReplyPlainPermanent(content string)

	// No functionality on interactions, check / cross reaction on messages
	Accept()
	Reject()

	HandleError(err error)
	HandleWarning(err error)

	// Utility functions
	Guild() (guild.Guild, error)
	Member() (member.Member, error)
	User() (user.User, error)
}
