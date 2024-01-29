package constant

import (
	"embed"
	"io/fs"
	"log"
	"strings"
)

var LatestVersion = map[string]string{}

func readFileContent(path string, f embed.FS) string {
	data, err := fs.ReadFile(f, path)
	if err != nil {
		return ""
	}
	return string(data)
}

func InitReleaseInfo(releaseInfo embed.FS) {
	version := strings.Trim(readFileContent(".release_version", releaseInfo), "\n\r")
	if version == "" {
		version = "default"
	}
	desc := strings.Trim(readFileContent(".release_log", releaseInfo), "\n\r")
	log.Printf("[message-nest] release version: %s", version)
	LatestVersion["version"] = version
	LatestVersion["desc"] = desc
}
