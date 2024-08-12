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

// setupFlags sets up the flags without parsing them
func setupFlags(fs *flag.FlagSet, cfg *Flags) {
	fs.IntVar(&cfg.DebugLevel, "debug", 0, "Debug level: 0=INFO, 1=DEBUG")
	fs.IntVar(&cfg.Concurrency, "concurrency", runtime.NumCPU(), "Number of concurrent operations")
	fs.StringVar(&cfg.ConfigFile, "config", "panorama.yaml", "Path to the Panorama configuration file")
	fs.StringVar(&cfg.SecretsFile, "secrets", ".secrets.yaml", "Path to the secrets file")
	fs.StringVar(&cfg.HostnameFilter, "filter", "", "Comma-separated list of hostname patterns to filter devices")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	fs.BoolVar(&cfg.NoPanorama, "nopanorama", false, "Use inventory.yaml instead of querying Panorama")
}

// ParseFlags parses command-line flags and returns a configuration object.
func ParseFlags() *Flags {
	cfg := &Flags{}
	setupFlags(flag.CommandLine, cfg)
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
	Result          string `json:"result,omitempty"`
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
	Inventory []InventoryDevice `yaml:"inventory"`
}

// InventoryDevice represents a single device in the inventory
type InventoryDevice struct {
	Hostname  string `yaml:"hostname"`
	IPAddress string `yaml:"ip_address"`
}

// Load reads configuration and secrets from YAML files and returns a Config struct.
// This function reads configuration data from a specified config file and secrets
// from a secrets file, combining them into a single Config struct.
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
func readYAMLFile(filename string, v interface{}) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	err = yaml.Unmarshal(data, v)
	if err != nil {
		return err
	}

	// If v is a pointer to a map[string]interface{}, convert nested maps
	if m, ok := v.(*map[string]interface{}); ok {
		*m = convertMap(*m)
	}

	return nil
}

// convertMap recursively converts map[interface{}]interface{} to map[string]interface{}
func convertMap(m map[string]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		switch v := v.(type) {
		case map[interface{}]interface{}:
			res[k] = convertMap(convertMapInterfaceToString(v))
		case []interface{}:
			res[k] = convertSlice(v)
		default:
			res[k] = v
		}
	}
	return res
}

// convertMapInterfaceToString converts map[interface{}]interface{} to map[string]interface{}
func convertMapInterfaceToString(m map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range m {
		res[fmt.Sprint(k)] = v
	}
	return res
}

// convertSlice recursively converts []interface{} elements
func convertSlice(s []interface{}) []interface{} {
	res := make([]interface{}, len(s))
	for i, v := range s {
		switch v := v.(type) {
		case map[interface{}]interface{}:
			res[i] = convertMap(convertMapInterfaceToString(v))
		case []interface{}:
			res[i] = convertSlice(v)
		default:
			res[i] = v
		}
	}
	return res
}
