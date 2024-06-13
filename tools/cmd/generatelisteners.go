package main

import (
	"bytes"
	_ "embed"
	_ "github.com/rxdn/gdl/gateway/payloads/events"
	"go/types"
	"golang.org/x/tools/go/packages"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

const (
	PackageName = "github.com/rxdn/gdl/gateway/payloads/events"
)

//go:embed listeners.tmpl
var listenersTemplate string

func main() {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg, PackageName)
	if err != nil {
		panic(err)
	}

	if len(pkgs) != 1 {
		panic("expected 1 package")
	}

	pkg := pkgs[0]
	scope := pkg.Types.Scope()

	events := make([]string, 0)

	for _, name := range scope.Names() {
		object := scope.Lookup(name)
		typeName := strings.TrimPrefix(object.Type().String(), PackageName+".")

		if typeName == object.Name() {
			if typeName == "EventBus" {
				continue
			}

			// Check if underlying type is a struct
			underlying := object.Type().Underlying()
			if _, ok := underlying.(*types.Struct); ok {
				events = append(events, object.Name())
			}
		}
	}

	tmpl, err := template.
		New("listeners").
		Funcs(template.FuncMap{
			"toScreamingSnakeCase": func(orig string) string {
				buf := strings.Builder{}
				for i, r := range orig {
					if i > 0 && 'A' <= r && r <= 'Z' {
						buf.WriteRune('_')
					}

					buf.WriteRune(unicode.ToUpper(r))
				}

				return buf.String()
			},
		}).
		Parse(listenersTemplate)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{
		"events": events,
	}); err != nil {
		panic(err)
	}

	path := filepath.Join(filepath.Dir("."), "listeners.go")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		panic(err)
	}
}
