package version

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestForkLatestReleaseFromGitHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/me/cli/releases/latest" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"tag_name":"0.19.0-fallback.2","name":"x","html_url":"https://example/rel"}`))
	}))
	defer srv.Close()

	oldAPI, oldRepo := forkGitHubAPI, ForkRepo
	forkGitHubAPI, ForkRepo = srv.URL, "me/cli"
	defer func() { forkGitHubAPI, ForkRepo = oldAPI, oldRepo }()

	rel, err := GetLatestRelease()
	if err != nil {
		t.Fatal(err)
	}
	if rel.TagName != "0.19.0-fallback.2" || rel.URL != "https://example/rel" {
		t.Fatalf("got %+v", rel)
	}
}

func TestForkVersionsCompare(t *testing.T) {
	c, err := CompareVersions("0.19.0-fallback.1", "0.19.0-fallback.2")
	if err != nil || c >= 0 {
		t.Fatalf("fallback.1 < fallback.2: c=%d err=%v", c, err)
	}
	c, err = CompareVersions("0.19.0-fallback.2", "0.20.0-fallback.1")
	if err != nil || c >= 0 {
		t.Fatalf("0.19 < 0.20: c=%d err=%v", c, err)
	}
}

func TestForkInstallScriptURL(t *testing.T) {
	old := ForkRepo
	ForkRepo = "me/cli"
	defer func() { ForkRepo = old }()
	if got := InstallScriptURL(); got != "https://raw.githubusercontent.com/me/cli/fallback/get-cli.sh" {
		t.Fatalf("got %q", got)
	}
}
