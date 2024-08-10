// config/config_test.go
package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Create temporary config and secrets files
	configContent := `
panorama:
  - hostname: test-panorama.example.com
`
	secretsContent := `
auth:
  panorama:
    username: panorama-user
    password: panorama-pass
  firewall:
    username: firewall-user
    password: firewall-pass
`
	tmpConfigFile := createTempFile(t, "config", configContent)
	defer func() {
		err := os.Remove(tmpConfigFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temporary config file: %v", err)
		}
	}()

	tmpSecretsFile := createTempFile(t, "secrets", secretsContent)
	defer func() {
		err := os.Remove(tmpSecretsFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temporary secrets file: %v", err)
		}
	}()

	// Test Load function
	config, err := Load(tmpConfigFile.Name(), tmpSecretsFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Check if the config is correctly loaded
	assert.Len(t, config.Panorama, 1)
	assert.Equal(t, "test-panorama.example.com", config.Panorama[0].Hostname)

	// Check if the secrets are correctly loaded
	assert.Equal(t, "panorama-user", config.Auth.Auth.Panorama.Username)
	assert.Equal(t, "panorama-pass", config.Auth.Auth.Panorama.Password)
	assert.Equal(t, "firewall-user", config.Auth.Auth.Firewall.Username)
	assert.Equal(t, "firewall-pass", config.Auth.Auth.Firewall.Password)
}

func TestLoadError(t *testing.T) {
	// Test with non-existent files
	_, err := Load("non-existent-config.yaml", "non-existent-secrets.yaml")
	assert.Error(t, err)
}

func TestReadYAMLFile(t *testing.T) {
	content := `
key1: value1
key2: value2
`
	tmpFile := createTempFile(t, "test", content)
	defer func() {
		err := os.Remove(tmpFile.Name())
		if err != nil {
			t.Errorf("Failed to remove temporary file: %v", err)
		}
	}()

	var result map[string]string
	err := readYAMLFile(tmpFile.Name(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result["key1"])
	assert.Equal(t, "value2", result["key2"])
}

func TestReadYAMLFileError(t *testing.T) {
	err := readYAMLFile("non-existent-file.yaml", &struct{}{})
	assert.Error(t, err)
}

// Helper function to create a temporary file with given content
func createTempFile(t *testing.T, prefix, content string) *os.File {
	tmpFile, err := os.CreateTemp("", prefix)
	assert.NoError(t, err)
	_, err = tmpFile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpFile.Close()
	assert.NoError(t, err)
	return tmpFile
}
