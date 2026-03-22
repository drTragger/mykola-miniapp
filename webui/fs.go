package webui

import (
	"embed"
	"io/fs"
)

//go:embed web/*
var embeddedFiles embed.FS

func staticFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "web")
}
