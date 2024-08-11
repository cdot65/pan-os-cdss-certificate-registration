// main_test.go
package main

import (
	"flag"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// Purpose: Tests the parseFlags function, ensuring that command-line arguments are correctly parsed and mapped to the config.Flags struct.
// Relevance: This test is crucial because it ensures that the program's configuration is correctly set up, affecting how the rest of the program behaves. The different test cases cover both default and custom values, which is essential for robust flag parsing validation.

func TestParseFlags(t *testing.T) {
	// Save original command-line arguments
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected *config.Flags
	}{
		{
			name: "Default values",
			args: []string{"cmd"},
			expected: &config.Flags{
				DebugLevel:     0,
				Concurrency:    runtime.NumCPU(), // Use actual number of CPUs
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
				"cmd",
				"-debug", "1",
				"-concurrency", "2",
				"-config", "custom.yaml",
				"-secrets", "custom_secrets.yaml",
				"-filter", "host1,host2",
				"-verbose",
				"-nopanorama",
			},
			expected: &config.Flags{
				DebugLevel:     1,
				Concurrency:    2,
				ConfigFile:     "custom.yaml",
				SecretsFile:    "custom_secrets.yaml",
				HostnameFilter: "host1,host2",
				Verbose:        true,
				NoPanorama:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set command-line arguments
			os.Args = tt.args

			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Call parseFlags
			result := config.ParseFlags()

			// Assert results
			assert.Equal(t, tt.expected.DebugLevel, result.DebugLevel)
			assert.Equal(t, tt.expected.Concurrency, result.Concurrency)
			assert.Equal(t, tt.expected.ConfigFile, result.ConfigFile)
			assert.Equal(t, tt.expected.SecretsFile, result.SecretsFile)
			assert.Equal(t, tt.expected.HostnameFilter, result.HostnameFilter)
			assert.Equal(t, tt.expected.Verbose, result.Verbose)
			assert.Equal(t, tt.expected.NoPanorama, result.NoPanorama)
		})
	}
}

//Purpose: Tests the printDeviceList function by capturing the output and checking the structure of the printed information.
//Relevance: The test focuses on validating the format and content of the device list printed to the console. This is important because the main.go file relies on correctly displaying the list of devices before registering WildFire, which provides critical feedback to the user.

func TestPrintDeviceList(t *testing.T) {
	l := logger.New(0, false)

	deviceList := []map[string]string{
		{"hostname": "device1", "ip": "192.168.1.1"},
		{"hostname": "device2", "ip": "192.168.1.2"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	utils.PrintDeviceList(deviceList, l)

	err := w.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("Failed to read captured output: %v", err)
	}
	output := string(out)

	// Check structure instead of exact content
	assert.Contains(t, output, "Device List:")
	assert.Contains(t, output, "Device 1:")
	assert.Contains(t, output, "Device 2:")
	assert.Contains(t, output, "hostname:")
	assert.Contains(t, output, "ip:")
}

//Purpose: Validates the structure of inventory.yaml by loading and unmarshaling the YAML file.
//Relevance: This test is relevant because main.go might use this inventory file when NoPanorama is true. Ensuring the structure is correct prevents runtime errors when accessing device information.

func TestInventoryYAMLStructure(t *testing.T) {
	data, err := os.ReadFile("inventory.yaml")
	assert.NoError(t, err)

	var inventory struct {
		Inventory []struct {
			Hostname  string `yaml:"hostname"`
			IPAddress string `yaml:"ip_address"`
		} `yaml:"inventory"`
	}

	err = yaml.Unmarshal(data, &inventory)
	assert.NoError(t, err)
	assert.NotEmpty(t, inventory.Inventory)
	for _, device := range inventory.Inventory {
		assert.NotEmpty(t, device.Hostname)
		assert.NotEmpty(t, device.IPAddress)
	}
}

//Purpose: Ensures that the panorama.yaml file has the correct structure by loading and unmarshaling it.
//Relevance: Since the main.go file depends on the panorama.yaml for configuration when querying Panorama, this test helps avoid issues related to malformed or incorrect configuration files.

func TestPanoramaYAMLStructure(t *testing.T) {
	data, err := os.ReadFile("panorama.yaml")
	assert.NoError(t, err)

	var panorama struct {
		Panorama []struct {
			Hostname string `yaml:"hostname"`
		} `yaml:"panorama"`
	}

	err = yaml.Unmarshal(data, &panorama)
	assert.NoError(t, err)
	assert.NotEmpty(t, panorama.Panorama)
	for _, p := range panorama.Panorama {
		assert.NotEmpty(t, p.Hostname)
	}
}

//Purpose: Validates the structure of the .secrets.yaml file to ensure that the required credentials are present and correctly structured.
//Relevance: The main.go file requires these credentials for authenticating with the Panorama and Firewall. This test is vital to ensure that authentication processes do not fail due to missing or incorrect credentials.

func TestSecretsYAMLStructure(t *testing.T) {
	data, err := os.ReadFile(".secrets.example.yaml")
	assert.NoError(t, err)

	var secrets struct {
		Auth struct {
			Panorama struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"panorama"`
			Firewall struct {
				Username string `yaml:"username"`
				Password string `yaml:"password"`
			} `yaml:"firewall"`
		} `yaml:"auth"`
	}

	err = yaml.Unmarshal(data, &secrets)
	assert.NoError(t, err)
	assert.NotEmpty(t, secrets.Auth.Panorama.Username)
	assert.NotEmpty(t, secrets.Auth.Panorama.Password)
	assert.NotEmpty(t, secrets.Auth.Firewall.Username)
	assert.NotEmpty(t, secrets.Auth.Firewall.Password)
}
