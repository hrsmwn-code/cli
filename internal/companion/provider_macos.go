//go:build darwin || linux || windows

package companion

import (
	"os"
	"path/filepath"
)

const darwinAppName = "Pangolin"

func macOSDesktopProvider(home, systemApplications string, secrets func(string) SecretsClient) Provider {
	dataDir := ""
	applications := []string{systemApplications}
	if home != "" {
		dataDir = filepath.Join(home, "Library", "Application Support", darwinAppName)
		applications = append(applications, filepath.Join(home, "Applications"))
	}
	for _, dir := range applications {
		info, err := os.Stat(filepath.Join(dir, darwinAppName+".app", "Contents", "Info.plist"))
		if err == nil && info.Mode().IsRegular() {
			return desktopProvider{defaultDataDir: dataDir, secrets: secrets}
		}
	}
	return nil
}
