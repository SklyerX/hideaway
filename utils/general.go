package utils

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func GetAppPaths() map[string]string {
	paths := make(map[string]string)

	userConfigDir, _ := os.UserConfigDir()
	userDataDir := filepath.Join(userConfigDir, ".hideaway")

	userHomeDir, _ := os.UserHomeDir()

	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	desktopDir := filepath.Join(userHomeDir, "Desktop")

	paths["userData"] = userDataDir
	paths["home"] = userHomeDir
	paths["exe"] = execDir
	paths["temp"] = os.TempDir()
	paths["desktop"] = desktopDir

	return paths
}

func CleanPath(path string) string {
	var result strings.Builder
	for _, r := range path {
		if unicode.IsSpace(r) {
			result.WriteRune(' ')
		} else if unicode.IsPrint(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}
