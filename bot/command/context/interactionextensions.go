package context

import (
	"github.com/rxdn/gdl/objects"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
)

type InteractionExtension struct {
	interaction interaction.ApplicationCommandInteraction
}

func NewInteractionExtension(interaction interaction.ApplicationCommandInteraction) InteractionExtension {
	return InteractionExtension{
		interaction: interaction,
	}
}

func (i InteractionExtension) Resolved() interaction.ResolvedData {
	return i.interaction.Data.Resolved
}

func (i InteractionExtension) ResolvedUser(id uint64) (user.User, bool) {
	user, ok := i.interaction.Data.Resolved.Users[objects.Snowflake(id)]
	return user, ok
}

func (i InteractionExtension) ResolvedMember(id uint64) (member.Member, bool) {
	member, ok := i.interaction.Data.Resolved.Members[objects.Snowflake(id)]
	return member, ok
}

func (i InteractionExtension) ResolvedRole(id uint64) (guild.Role, bool) {
	role, ok := i.interaction.Data.Resolved.Roles[objects.Snowflake(id)]
	return role, ok
}

func (i InteractionExtension) ResolvedChannel(id uint64) (channel.Channel, bool) {
	channel, ok := i.interaction.Data.Resolved.Channels[objects.Snowflake(id)]
	return channel, ok
}

func (i InteractionExtension) ResolvedMessage(id uint64) (message.Message, bool) {
	message, ok := i.interaction.Data.Resolved.Messages[objects.Snowflake(id)]
	return message, ok
}

func (i InteractionExtension) ResolvedAttachment(id uint64) (channel.Attachment, bool) {
	attachment, ok := i.interaction.Data.Resolved.Attachments[objects.Snowflake(id)]
	return attachment, ok
}
