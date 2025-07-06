package utils

import (
	"os"
	"path/filepath"
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
