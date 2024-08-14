package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"runtime"
	"testing"
)

func createTestFlags() *Flags {
	return &Flags{
		DebugLevel:     0,
		Concurrency:    runtime.NumCPU(),
		ConfigFile:     "panorama.yaml",
		SecretsFile:    ".secrets.yaml",
		HostnameFilter: "",
		Verbose:        false,
		NoPanorama:     false,
	}
}

func TestSetupFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *Flags
	}{
		{
			name: "Default values",
			args: []string{},
			expected: &Flags{
				DebugLevel:     0,
				Concurrency:    runtime.NumCPU(),
				ConfigFile:     "panorama.yaml",
				SecretsFile:    ".secrets.yaml",
				HostnameFilter: "",
				Verbose:        false,
				NoPanorama:     false,
			},
		},
		{
			name: "Custom values",
			args: []string{
				"-debug", "1",
				"-concurrency", "4",
				"-config", "custom.yaml",
				"-secrets", "custom_secrets.yaml",
				"-filter", "fw-*",
				"-verbose",
				"-nopanorama",
			},
			expected: &Flags{
				DebugLevel:     1,
				Concurrency:    4,
				ConfigFile:     "custom.yaml",
				SecretsFile:    "custom_secrets.yaml",
				HostnameFilter: "fw-*",
				Verbose:        true,
				NoPanorama:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			cfg := &Flags{}
			setupFlags(fs, cfg)

			err := fs.Parse(tt.args)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, cfg)
		})
	}
}
