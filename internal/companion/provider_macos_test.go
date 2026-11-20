//go:build darwin || linux || windows

package companion

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMacOSProviderLeavesStandaloneCLIAvailableWithoutDesktopApp(t *testing.T) {
	home := t.TempDir()
	applications := t.TempDir()
	// App data left behind by an uninstall must not take over CLI auth.
	dataDir := filepath.Join(home, "Library", "Application Support", "Pangolin")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if got := macOSDesktopProvider(home, applications, nil); got != nil {
		t.Fatal("companion provider enabled without an installed desktop app")
	}
}

func TestMacOSProviderFindsSystemAndUserApplications(t *testing.T) {
	for _, perUser := range []bool{false, true} {
		home := t.TempDir()
		applications := t.TempDir()
		installDir := applications
		if perUser {
			installDir = filepath.Join(home, "Applications")
		}
		contents := filepath.Join(installDir, "Pangolin.app", "Contents")
		if err := os.MkdirAll(contents, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(contents, "Info.plist"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		got := macOSDesktopProvider(home, applications, nil)
		if got == nil || got.DefaultDataDir() != filepath.Join(home, "Library", "Application Support", "Pangolin") {
			t.Fatalf("per-user install %v: provider = %#v", perUser, got)
		}
	}
}
