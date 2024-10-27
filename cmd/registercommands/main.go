package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/TicketsBot/worker/bot/command/manager"
	"github.com/TicketsBot/worker/i18n"
	"github.com/rxdn/gdl/objects/interaction"
	"github.com/rxdn/gdl/rest"
)

var (
	Token         = flag.String("token", "", "Bot token to create commands for")
	ApplicationId = flag.Uint64("id", 508391840525975553, "Application ID")
	GuildId       = flag.Uint64("guild", 0, "Guild to create the commands for")

	AdminCommandGuildId = flag.Uint64("admin-guild", 0, "Guild to create the admin commands in")
	MergeGuildCommands  = flag.Bool("merge", true, "Don't overwrite existing commands")
)

func main() {
	flag.Parse()
	if *Token == "" {
		panic("no token")
	}

	i18n.Init()

	commandManager := new(manager.CommandManager)
	commandManager.RegisterCommands()

	data, adminCommands := commandManager.BuildCreatePayload(false, AdminCommandGuildId)

	var err error
	if *GuildId == 0 {
		must(rest.ModifyGlobalCommands(context.Background(), *Token, nil, *ApplicationId, data))
	} else {
		must(rest.ModifyGuildCommands(context.Background(), *Token, nil, *ApplicationId, *GuildId, data))
	}

	if err != nil {
		panic(err)
	}

	if AdminCommandGuildId != nil && *AdminCommandGuildId != 0 {
		if MergeGuildCommands != nil && *MergeGuildCommands {
			cmds := must(rest.GetGuildCommands(context.Background(), *Token, nil, *ApplicationId, *AdminCommandGuildId))
			for _, cmd := range cmds {
				var found bool
				for _, newCmd := range adminCommands {
					if cmd.Name == newCmd.Name {
						found = true
						break
					}
				}

				if !found {
					adminCommands = append(adminCommands, rest.CreateCommandData{
						Id:          cmd.Id,
						Name:        cmd.Name,
						Description: cmd.Description,
						Options:     cmd.Options,
						Type:        interaction.ApplicationCommandTypeChatInput,
					})
				}
			}
		}

		must(rest.ModifyGuildCommands(context.Background(), *Token, nil, *ApplicationId, *AdminCommandGuildId, adminCommands))
	}

	cmds := must(rest.GetGlobalCommands(context.Background(), *Token, nil, *ApplicationId))
	marshalled := must(json.MarshalIndent(cmds, "", "    "))

	fmt.Println(string(marshalled))
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}
