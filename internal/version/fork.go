package version

// Fork-only: the TCP 443 fallback build checks for updates against the
// fork's own GitHub releases and installs from the fork's install script,
// so `pangolin update` never replaces it with an upstream build.

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ForkRepo is the GitHub repo ("owner/name") releases come from. Override
// at build time: -ldflags "-X github.com/fosrl/cli/internal/version.ForkRepo=owner/cli"
var ForkRepo = "hrsmwn-code/cli"

// ForkBranch holds the fork's install script.
var ForkBranch = "fallback"

var forkGitHubAPI = "https://api.github.com"

// InstallScriptURL is the fork's get-cli.sh.
func InstallScriptURL() string {
	return "https://raw.githubusercontent.com/" + ForkRepo + "/" + ForkBranch + "/get-cli.sh"
}

func getForkLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, forkGitHubAPI+"/repos/"+ForkRepo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "pangolin-cli")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("failed to fetch release: status %d, body: %s", resp.StatusCode, body)
	}
	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to parse release: %w", err)
	}
	if rel.TagName == "" {
		return nil, fmt.Errorf("release has no tag")
	}
	return &rel, nil
}
