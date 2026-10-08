package constant

import (
	"embed"
	"io/fs"
	"strings"

	"message-nest/pkg/logging"
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
	logging.App.Infof("发布版本: %s", version)
	LatestVersion["version"] = version
	LatestVersion["desc"] = desc
}
