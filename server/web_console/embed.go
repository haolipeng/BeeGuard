package web_console

import "embed"

//go:embed dist/*
var StaticFS embed.FS
