package web

import (
	"embed"
	"io/fs"
)

//go:embed dist dist/*
var embeddedFiles embed.FS

func StaticFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist")
}
