package main

import (
	"bytes"
	_ "embed"
	"github.com/TicketsBot/worker/bot/command"
	"github.com/TicketsBot/worker/bot/command/manager"
	"github.com/TicketsBot/worker/bot/command/registry"
	"github.com/TicketsBot/worker/bot/utils"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/interaction"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"text/template"
)

type executorData struct {
	ImportName string
	Arguments  []command.Argument
}

//go:embed cmdcaller.tmpl
var callerTemplate string

var typeMap = map[interaction.ApplicationCommandOptionType]reflect.Type{
	interaction.OptionTypeString:      reflect.TypeOf(""),                   // string
	interaction.OptionTypeInteger:     reflect.TypeOf(int(0)),               // int
	interaction.OptionTypeBoolean:     reflect.TypeOf(false),                // bool
	interaction.OptionTypeUser:        reflect.TypeOf(uint64(0)),            // snowflake
	interaction.OptionTypeChannel:     reflect.TypeOf(uint64(0)),            // snowflake
	interaction.OptionTypeRole:        reflect.TypeOf(uint64(0)),            // snowflake
	interaction.OptionTypeMentionable: reflect.TypeOf(uint64(0)),            // snowflake
	interaction.OptionTypeNumber:      reflect.TypeOf(float64(0)),           // float64
	interaction.OptionTypeAttachment:  reflect.TypeOf(channel.Attachment{}), // attachment
}

func main() {
	cm := manager.CommandManager{}
	cm.RegisterCommands()

	allCmds := make([]registry.Command, 0, len(cm.GetCommands()))
	for _, cmd := range cm.GetCommands() {
		allCmds = append(allCmds, cmd)
		for _, sub := range cmd.Properties().Children {
			allCmds = append(allCmds, sub)
		}
	}

	var packagePaths []string
	var executors []executorData

	for _, cmd := range allCmds {
		t := reflect.TypeOf(cmd)
		pkg := t.PkgPath()
		if !utils.Contains(packagePaths, pkg) {
			packagePaths = append(packagePaths, pkg)
		}

		importName := pkg[strings.LastIndex(pkg, "/")+1:] + "." + t.Name()

		executors = append(executors, executorData{
			ImportName: importName,
			Arguments:  cmd.Properties().Arguments,
		})
	}

	// Order executors by import name for reproducible builds
	sort.Slice(executors, func(i, j int) bool {
		return executors[i].ImportName < executors[j].ImportName
	})

	tmpl, err := template.New("caller").
		Funcs(template.FuncMap{
			"panic": func(msg string) string {
				panic(msg)
			},
		}).
		Parse(callerTemplate)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{
		"imports":   packagePaths,
		"executors": executors,
		"typeMap":   typeMap,
	}); err != nil {
		panic(err)
	}

	path := filepath.Join(filepath.Dir("."), "caller.go")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		panic(err)
	}
}
