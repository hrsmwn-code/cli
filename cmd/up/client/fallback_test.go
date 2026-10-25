package client

import (
	"reflect"
	"testing"

	"github.com/fosrl/newt/fallback"
	"github.com/spf13/cobra"
)

func newFallbackCmd() (*cobra.Command, *fallbackOpts) {
	cmd := &cobra.Command{Use: "t", Run: func(*cobra.Command, []string) {}}
	f := &fallbackOpts{}
	f.register(cmd)
	return cmd, f
}

func TestFallbackDefaults(t *testing.T) {
	cmd, f := newFallbackCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	s, err := f.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s != fallback.DefaultSettings() {
		t.Fatalf("got %+v", s)
	}
	if args := f.subprocessArgs(); len(args) != 0 {
		t.Fatalf("defaults must not be forwarded, got %v", args)
	}
}

func TestFallbackFlagsForwardedToSubprocess(t *testing.T) {
	cmd, f := newFallbackCmd()
	if err := cmd.ParseFlags([]string{"--transport", "tcp", "--fallback-relay-path", "/x", "--fallback-echo-port", "9"}); err != nil {
		t.Fatal(err)
	}
	s, err := f.settings()
	if err != nil {
		t.Fatal(err)
	}
	if s.Mode != fallback.ModeTCP || s.RelayPath != "/x" || s.EchoPort != 9 {
		t.Fatalf("got %+v", s)
	}
	want := []string{"--transport", "tcp", "--fallback-relay-path", "/x", "--fallback-echo-port", "9"}
	if got := f.subprocessArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("subprocess args %v, want %v", got, want)
	}
}

func TestFallbackEnvUsedWhenFlagUnset(t *testing.T) {
	t.Setenv(fallback.EnvTransport, "tcp")
	cmd, f := newFallbackCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	if err := f.applyEnv(cmd); err != nil {
		t.Fatal(err)
	}
	if s, _ := f.settings(); s.Mode != fallback.ModeTCP {
		t.Fatalf("env not applied: %+v", s)
	}

	cmd, f = newFallbackCmd()
	if err := cmd.ParseFlags([]string{"--transport", "udp"}); err != nil {
		t.Fatal(err)
	}
	if err := f.applyEnv(cmd); err != nil {
		t.Fatal(err)
	}
	if s, _ := f.settings(); s.Mode != fallback.ModeUDP {
		t.Fatalf("flag must win over env: %+v", s)
	}
}

func TestFallbackConfigUsedWhenFlagAndEnvUnset(t *testing.T) {
	cmd, f := newFallbackCmd()
	if err := cmd.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	f.applyConfig(cmd, mapConfig{"up.transport": "tcp", "up.fallback_relay_path": "/cfg"})
	s, _ := f.settings()
	if s.Mode != fallback.ModeTCP || s.RelayPath != "/cfg" {
		t.Fatalf("config not applied: %+v", s)
	}
	// config-sourced values must reach the root subprocess, which can't read the user's config
	want := []string{"--transport", "tcp", "--fallback-relay-path", "/cfg"}
	if got := f.subprocessArgs(); !reflect.DeepEqual(got, want) {
		t.Fatalf("subprocess args %v, want %v", got, want)
	}
}

func TestFallbackInvalidTransport(t *testing.T) {
	cmd, f := newFallbackCmd()
	if err := cmd.ParseFlags([]string{"--transport", "quic"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.settings(); err == nil {
		t.Fatal("invalid transport must be rejected")
	}
}

type mapConfig map[string]string

func (m mapConfig) IsSet(k string) bool       { _, ok := m[k]; return ok }
func (m mapConfig) GetString(k string) string { return m[k] }
