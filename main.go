package main

import (
	"embed"
	"message-nest/bootstrap"
)

var (
	//go:embed web/dist/*
	f embed.FS

	//go:embed .release*
	rf embed.FS
)

func main() {
	bootstrap.New(f, rf).Run()
}
