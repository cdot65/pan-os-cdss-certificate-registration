package devices

import (
	"gopkg.in/yaml.v2"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPanosClient is a mock implementation of the PanosClient interface
type MockPanosClient struct {
	mock.Mock
}

func (m *MockPanosClient) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPanosClient) Op(cmd interface{}, vsys string, extras interface{}, ans interface{}) ([]byte, error) {
	args := m.Called(cmd, vsys, extras, ans)
	return args.Get(0).([]byte), args.Error(1)
}

func setupTestConfig() (*config.Config, error) {
	// Sample configuration
	configYaml := `
panorama:
  - hostname: test-panorama.example.com
`
	secretsYaml := `
auth:
  credentials:
    panorama:
      username: test-user
      password: test-pass
    firewall:
      username: fw-user
      password: fw-pass
`

	cfg := &config.Config{}
	if err := yaml.Unmarshal([]byte(configYaml), cfg); err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal([]byte(secretsYaml), &cfg.Auth); err != nil {
		return nil, err
	}

	return cfg, nil
}

func TestNewDeviceManager(t *testing.T) {
	conf, err := setupTestConfig()
	assert.NoError(t, err)
	l := logger.New(0, false)

	dm := NewDeviceManager(conf, l)

	assert.NotNil(t, dm)
	assert.Equal(t, conf, dm.config)
	assert.Equal(t, l, dm.logger)
	assert.Nil(t, dm.panosClientFactory)
}

func TestSetNgfwWorkflow(t *testing.T) {
	conf := &config.Config{}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	dm.SetNgfwWorkflow()
	assert.NotNil(t, dm.panosClientFactory)

	// Test the factory creates a PanosClient
	client := dm.panosClientFactory("test", "user", "pass")
	assert.NotNil(t, client)
}

func TestSetPanoramaWorkflow(t *testing.T) {
	conf := &config.Config{}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	dm.SetPanoramaWorkflow()
	assert.NotNil(t, dm.panosClientFactory)

	// Test the factory creates a PanosClient
	client := dm.panosClientFactory("test", "user", "pass")
	assert.NotNil(t, client)
}
