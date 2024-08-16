package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name           string
		configContent  string
		secretsContent string
		flags          *Flags
		expectedConfig *Config
		expectError    bool
	}{
		{
			name: "Valid config and secrets",
			configContent: `
panorama:
  - hostname: test-panorama.example.com
`,
			secretsContent: `
auth:
  panorama:
    username: panorama-user
    password: panorama-pass
  firewall:
    username: firewall-user
    password: firewall-pass
`,
			flags: createTestFlags(),
			expectedConfig: &Config{
				Panorama: []struct {
					Hostname string `yaml:"hostname"`
				}{
					{Hostname: "test-panorama.example.com"},
				},
				Auth: AuthConfig{
					Credentials: struct {
						Panorama struct {
							Username string `yaml:"username"`
							Password string `yaml:"password"`
						} `yaml:"panorama"`
						Firewall struct {
							Username string `yaml:"username"`
							Password string `yaml:"password"`
						} `yaml:"firewall"`
					}{
						Panorama: struct {
							Username string `yaml:"username"`
							Password string `yaml:"password"`
						}{
							Username: "panorama-user",
							Password: "panorama-pass",
						},
						Firewall: struct {
							Username string `yaml:"username"`
							Password string `yaml:"password"`
						}{
							Username: "firewall-user",
							Password: "firewall-pass",
						},
					},
				},
				HostnameFilter: "",
			},
			expectError: false,
		},
		{
			name:           "Invalid config file",
			configContent:  "invalid: yaml: content",
			secretsContent: "",
			expectedConfig: nil,
			expectError:    true,
		},
		{
			name:           "Invalid secrets file",
			configContent:  "panorama: []",
			secretsContent: "invalid: yaml: content",
			expectedConfig: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := createTempFile(t, "config", tt.configContent)
			defer func() {
				err := os.Remove(configFile.Name())
				assert.NoError(t, err)
			}()

			secretsFile := createTempFile(t, "secrets", tt.secretsContent)
			defer func() {
				err := os.Remove(secretsFile.Name())
				assert.NoError(t, err)
			}()

			config, err := Load(configFile.Name(), secretsFile.Name(), tt.flags)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, reflect.DeepEqual(tt.expectedConfig, config), "Configs do not match")
			}
		})
	}
}

func TestLoadError(t *testing.T) {
	flags := createTestFlags()
	_, err := Load("non-existent-config.yaml", "non-existent-secrets.yaml", flags)
	assert.Error(t, err)
}

func TestReadYAMLFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expected    map[string]interface{}
		expectError bool
	}{
		{
			name: "Valid YAML",
			content: `
key1: value1
key2:
  nested1: nestedvalue1
  nested2: nestedvalue2
`,
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"nested1": "nestedvalue1",
					"nested2": "nestedvalue2",
				},
			},
			expectError: false,
		},
		{
			name:        "Invalid YAML",
			content:     "key1: value1\nkey2: : invalid",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, "test", tt.content)
			defer func() {
				err := os.Remove(tmpFile.Name())
				assert.NoError(t, err)
			}()

			var result map[string]interface{}
			err := readYAMLFile(tmpFile.Name(), &result)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReadYAMLFileError(t *testing.T) {
	err := readYAMLFile("non-existent-file.yaml", &struct{}{})
	assert.Error(t, err)
}

// Helper function to create a temporary file with given content
func createTempFile(t *testing.T, prefix, content string) *os.File {
	tmpFile, err := os.CreateTemp("", prefix)
	require.NoError(t, err)

	_, err = tmpFile.Write([]byte(content))
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile
}
