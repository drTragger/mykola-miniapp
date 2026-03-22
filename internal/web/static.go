package web

import (
	"embed"
	"io/fs"
)

//go:embed web/*
var embeddedFiles embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "web")
}
