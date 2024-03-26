package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	lua "github.com/yuin/gopher-lua"
)

//go:embed stub.lua
var stubScript string

//go:embed table.html
var tableTemplate string

type FormatterData struct {
	Name        string
	Url         string
	Description string
}

type TemplateData struct {
	Data []FormatterData
}

func makePathToInit(pathStr string) string {
	if filepath.IsAbs(pathStr) {
		return pathStr
	}
	return fmt.Sprintf("./%s", pathStr)
}

func main() {
	basePath := "conform.nvim"
	formattersDir := filepath.Join(basePath, "lua", "conform", "formatters")
	entries, err := os.ReadDir(formattersDir)
	if err != nil {
		log.Fatal(err)
	}
	l := lua.NewState()
	defer l.Close()
	l.DoString(fmt.Sprintf("package.path = '%s/?.lua;' .. package.path", makePathToInit(filepath.Join(basePath, "lua"))))

	l.DoString(stubScript)
	refs := []FormatterData{}
	for _, e := range entries {
		trimmedName := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if trimmedName == "init" {
			continue
		}
		if err := l.DoFile(filepath.Join(formattersDir, e.Name())); err != nil {
			log.Fatal(err)
		}
		lv := l.Get(-1)
		if tbl, ok := lv.(*lua.LTable); ok {
			lv := l.GetTable(tbl, lua.LString("meta"))
			if meta, ok := lv.(*lua.LTable); ok {
				isDeprecated := l.GetTable(meta, lua.LString("deprecated"))
				if isDeprecated.Type() == lua.LTBool {
					continue
				}
				formatterUrl := l.GetTable(meta, lua.LString("url"))
				formatterDescription := l.GetTable(meta, lua.LString("description"))
				if formatterUrl.Type() == lua.LTString && formatterDescription.Type() == lua.LTString {
					refs = append(refs, FormatterData{Name: trimmedName, Url: formatterUrl.String(), Description: formatterDescription.String()})
				}
			}
		}
		l.Pop(-1)
	}
	tmpl := template.Must(template.New("").Parse(tableTemplate))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, TemplateData{Data: refs}); err != nil {
		log.Fatal(err)
	}
	converter := md.NewConverter("", true, nil)
	converter.Use(plugin.GitHubFlavored())
	markdown, err := converter.ConvertString(buf.String())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(markdown)
}
