//go:build darwin

package companion

import "github.com/fosrl/cli/internal/config"

func platformProvider() Provider {
	// Use the original user's home when the CLI runs under sudo.
	home, _ := config.UserHomeDir()
	return macOSDesktopProvider(home, "/Applications", newSocketSecretsClient)
}
