package client

// Fork-only: TCP 443 fallback options for `pangolin up client`.
// Precedence: flag > env (PANGOLIN_TRANSPORT, ...) > config (up.transport,
// ...) > default. client.go only calls the methods below.

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fosrl/newt/fallback"
	"github.com/spf13/cobra"
)

const (
	flagTransport = "transport"
	flagRelayPath = "fallback-relay-path"
	flagEchoPort  = "fallback-echo-port"
)

type configGetter interface {
	IsSet(key string) bool
	GetString(key string) string
}

type fallbackOpts struct {
	transport string
	relayPath string
	echoPort  int

	// fromConfig marks values that came from the persisted config; they
	// must be forwarded to the detached (root) subprocess explicitly.
	fromConfig map[string]bool
}

func (f *fallbackOpts) register(cmd *cobra.Command) {
	d := fallback.DefaultSettings()
	cmd.Flags().StringVar(&f.transport, flagTransport, string(d.Mode), "Transport to the Pangolin server: auto (UDP, fall back to TCP 443 when blocked), tcp (always TCP 443), udp (never fall back). Env "+fallback.EnvTransport)
	cmd.Flags().StringVar(&f.relayPath, flagRelayPath, d.RelayPath, "URL path of the TCP 443 fallback relay. Env "+fallback.EnvRelayPath)
	cmd.Flags().IntVar(&f.echoPort, flagEchoPort, d.EchoPort, "UDP echo port used by auto mode to detect whether UDP works. Env "+fallback.EnvEchoPort)
}

func (f *fallbackOpts) applyEnv(cmd *cobra.Command) error {
	for flag, env := range map[string]string{
		flagTransport: fallback.EnvTransport,
		flagRelayPath: fallback.EnvRelayPath,
		flagEchoPort:  fallback.EnvEchoPort,
	} {
		if cmd.Flags().Changed(flag) {
			continue
		}
		if v := os.Getenv(env); v != "" {
			if err := cmd.Flags().Set(flag, v); err != nil {
				return fmt.Errorf("invalid %s value %q: %w", env, v, err)
			}
		}
	}
	return nil
}

func (f *fallbackOpts) applyConfig(cmd *cobra.Command, cfg configGetter) {
	if cfg == nil {
		return
	}
	f.fromConfig = map[string]bool{}
	for flag, key := range map[string]string{
		flagTransport: "up.transport",
		flagRelayPath: "up.fallback_relay_path",
		flagEchoPort:  "up.fallback_echo_port",
	} {
		if cmd.Flags().Changed(flag) || !cfg.IsSet(key) {
			continue
		}
		if err := cmd.Flags().Set(flag, cfg.GetString(key)); err == nil {
			f.fromConfig[flag] = true
		}
	}
}

func (f *fallbackOpts) settings() (fallback.Settings, error) {
	m, err := fallback.ParseMode(f.transport)
	if err != nil {
		return fallback.Settings{}, err
	}
	s := fallback.Settings{Mode: m, RelayPath: f.relayPath, EchoPort: f.echoPort}
	return s, s.Validate()
}

// subprocessArgs returns the flags to forward to the detached subprocess:
// every value that differs from the default.
func (f *fallbackOpts) subprocessArgs() []string {
	d := fallback.DefaultSettings()
	var args []string
	if f.transport != "" && f.transport != string(d.Mode) {
		args = append(args, "--"+flagTransport, f.transport)
	}
	if f.relayPath != "" && f.relayPath != d.RelayPath {
		args = append(args, "--"+flagRelayPath, f.relayPath)
	}
	if f.echoPort != 0 && f.echoPort != d.EchoPort {
		args = append(args, "--"+flagEchoPort, strconv.Itoa(f.echoPort))
	}
	return args
}
