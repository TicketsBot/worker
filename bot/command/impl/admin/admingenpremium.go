package admin

import (
	"fmt"
	"github.com/TicketsBot/common/permission"
	"github.com/TicketsBot/common/sentry"
	database "github.com/TicketsBot/database/translations"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/dbclient"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/gofrs/uuid"
	"strconv"
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
	}
}

func (AdminGenPremiumCommand) Execute(ctx command.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		return
	}

	days, err := strconv.Atoi(ctx.Args[0]); if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}

	amount := 1
	if len(ctx.Args) == 2 {
		if a, err := strconv.Atoi(ctx.Args[1]); err == nil {
			amount = a
		}
	}

	keys := make([]string, 0)
	for i := 0; i < amount; i++ {
		key, err := uuid.NewV4()
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			continue
		}

		err = dbclient.Client.PremiumKeys.Create(key, time.Hour * 24 * time.Duration(days))
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			keys = append(keys, key.String())
		}
	}

	dmChannel, err := ctx.Worker.CreateDM(ctx.Author.Id); if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}

	content := "```"
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}
	content = strings.TrimSuffix(content, "\n")
	content += "```"

	_, err = ctx.Worker.CreateMessage(dmChannel.Id, content); if err != nil {
		ctx.SendEmbedRaw(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}

	ctx.ReactWithCheck()
}
