package main

import (
	"embed"
	"message-nest/bootstrap"
)

var (
	// Include underscore-prefixed Vite helper chunks in the embedded frontend.
	//go:embed all:web/dist
	f embed.FS

	//go:embed .release*
	rf embed.FS
)

func main() {
	bootstrap.New(f, rf).Run()
}
