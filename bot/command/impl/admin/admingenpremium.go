package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/premium"
	"github.com/TicketsBot/common/sentry"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/gofrs/uuid"
	"github.com/rxdn/gdl/objects/interaction"
	"strings"
	"time"
)

type AdminGenPremiumCommand struct {
}

func (AdminGenPremiumCommand) Properties() command.Properties {
	return command.Properties{
		Name:            "genpremium",
		Description:     database.HelpAdminGenPremium,
		Aliases:         []string{"gp", "gk", "generatepremium", "genkeys", "generatekeys"},
		PermissionLevel: permission.Everyone,
		Category:        command.Settings,
		AdminOnly:       true,
		MessageOnly:     true,
		Arguments: command.Arguments(
			command.NewRequiredArgument("length", "Length in days of the key", interaction.OptionTypeInteger, database.MessageInvalidArgument),
			command.NewOptionalArgument("amount", "Amount of keys to generate", interaction.OptionTypeInteger, database.MessageInvalidArgument),
			command.NewOptionalArgument("whitelabel", "Should the keys be for premium or whitelabel", interaction.OptionTypeBoolean, database.MessageInvalidArgument),
		),
	}
}

func (c AdminGenPremiumCommand) GetExecutor() interface{} {
	return c.Execute
}

func (AdminGenPremiumCommand) Execute(ctx command.CommandContext, length int, amountRaw *int, whitelabel *bool) {
	amount := 1
	if amountRaw != nil {
		amount = *amountRaw
	}

	tier := premium.Premium
	if whitelabel != nil && *whitelabel  {
		tier = premium.Whitelabel
	}

	keys := make([]string, 0)
	for i := 0; i < amount; i++ {
		key, err := uuid.NewV4()
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			continue
		}

		err = dbclient.Client.PremiumKeys.Create(key, time.Hour*24*time.Duration(length), int(tier))
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			keys = append(keys, key.String())
		}
	}

	dmChannel, err := ctx.Worker().CreateDM(ctx.UserId())
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Admin", err.Error())
		ctx.Reject()
		return
	}

	content := "```"
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}
	content = strings.TrimSuffix(content, "\n")
	content += "```"

	_, err = ctx.Worker().CreateMessage(dmChannel.Id, content)
	if err != nil {
		ctx.ReplyRaw(utils.Red, "Admin", err.Error())
		ctx.Reject()
		return
	}

	ctx.Accept()
}
