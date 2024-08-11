// Package config/config.go
package config

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

// Flags represents the command-line flags
type Flags struct {
	DebugLevel     int
	Concurrency    int
	ConfigFile     string
	SecretsFile    string
	HostnameFilter string
	Verbose        bool
	NoPanorama     bool
}

// ParseFlags parses command-line flags and returns a configuration object.
// This function sets up and parses command-line flags for various configuration options,
// including debug level, concurrency, file paths, and operational modes.
func ParseFlags() *Flags {
	cfg := &Flags{}
	flag.IntVar(&cfg.DebugLevel, "debug", 0, "Debug level: 0=INFO, 1=DEBUG")
	flag.IntVar(&cfg.Concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent operations")
	flag.StringVar(&cfg.ConfigFile, "config", "panorama.yaml", "Path to the Panorama configuration file")
	flag.StringVar(&cfg.SecretsFile, "secrets", ".secrets.yaml", "Path to the secrets file")
	flag.StringVar(&cfg.HostnameFilter, "filter", "", "Comma-separated list of hostname patterns to filter devices")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&cfg.NoPanorama, "nopanorama", false, "Use inventory.yaml instead of querying Panorama")
	flag.Parse()
	return cfg
}

// Panorama represents the configuration details for Panorama.
type Panorama struct {
	Hostname string `yaml:"hostname"`
}

// Config represents the overall configuration containing Panorama details.
type Config struct {
	Panorama []struct {
		Hostname string `yaml:"hostname"`
	} `yaml:"panorama"`
	Auth AuthConfig
}

// AuthConfig represents the authentication configuration.
type AuthConfig struct {
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

// DeviceEntry represents a single device entry from the Panorama response.
type DeviceEntry struct {
	Name            string `xml:"name,attr"`
	Serial          string `xml:"serial"`
	Hostname        string `xml:"hostname"`
	IPAddress       string `xml:"ip-address"`
	IPv6Address     string `xml:"ipv6-address"`
	Model           string `xml:"model"`
	SWVersion       string `xml:"sw-version"`
	AppVersion      string `xml:"app-version"`
	AVVersion       string `xml:"av-version"`
	WildfireVersion string `xml:"wildfire-version"`
	ThreatVersion   string `xml:"threat-version"`
}

// DevicesResponse represents the structure of the XML response from Panorama.
type DevicesResponse struct {
	XMLName xml.Name `xml:"response"`
	Status  string   `xml:"status,attr"`
	Result  struct {
		Devices struct {
			Entries []DeviceEntry `xml:"entry"`
		} `xml:"devices"`
	} `xml:"result"`
}

// Inventory represents the structure of the inventory.yaml file
type Inventory struct {
	Inventory []struct {
		Hostname  string `yaml:"hostname"`
		IPAddress string `yaml:"ip_address"`
	} `yaml:"inventory"`
}

// Load reads configuration and secrets from YAML files and returns a Config struct.

// This function reads configuration data from a specified config file and secrets
// from a secrets file, combining them into a single Config struct.

// Attributes:
//   configFile (string): Path to the main configuration YAML file.
//   secretsFile (string): Path to the secrets YAML file.

// Error:
//   error: If there's an issue reading either the config or secrets file.

// Return:
//   *Config: Pointer to the populated Config struct.
//   error: Nil if successful, otherwise an error describing what went wrong.

func Load(configFile, secretsFile string) (*Config, error) {
	var config Config
	if err := readYAMLFile(configFile, &config); err != nil {
		return nil, fmt.Errorf("failed to read Panorama config: %w", err)
	}
	if err := readYAMLFile(secretsFile, &config.Auth); err != nil {
		return nil, fmt.Errorf("failed to read secrets: %w", err)
	}
	return &config, nil
}

// readYAMLFile reads and unmarshals YAML data from a file into a provided interface.

// This function reads the contents of a YAML file specified by the filename,
// and unmarshals the data into the provided interface.

// Attributes:
//   filename (string): The path to the YAML file to be read.
//   v (interface{}): A pointer to the variable where the unmarshaled data will be stored.

// Error:
//   error: An error is returned if the file cannot be read or if the YAML data cannot be unmarshaled.

// Return:
//   error: nil if successful, otherwise an error describing what went wrong.

func readYAMLFile(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return yaml.Unmarshal(data, v)
}
