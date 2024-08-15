package devices

import (
	"github.com/PaloAltoNetworks/pango"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPanoramaClient is a mock implementation of the PanosClient interface
type MockPanoramaClient struct {
	mock.Mock
}

func (m *MockPanoramaClient) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPanoramaClient) Op(cmd interface{}, vsys string, extras interface{}, ans interface{}) ([]byte, error) {
	args := m.Called(cmd, vsys, extras, ans)
	return args.Get(0).([]byte), args.Error(1)
}

func TestDefaultPanoramaClientFactory(t *testing.T) {
	client := defaultPanoramaClientFactory("test-host", "test-user", "test-pass")
	assert.NotNil(t, client)
	assert.IsType(t, &pango.Panorama{}, client)
}

func TestGetDevicesFromPanorama(t *testing.T) {
	// Setup
	conf := &config.Config{
		Panorama: []struct {
			Hostname string `yaml:"hostname"`
		}{
			{Hostname: "test-panorama"},
		},
		Auth: config.AuthConfig{
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
					Username: "test-user",
					Password: "test-pass",
				},
			},
		},
	}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	mockClient := new(MockPanoramaClient)
	dm.panosClientFactory = func(hostname, username, password string) PanosClient {
		return mockClient
	}

	// Mock the Initialize method
	mockClient.On("Initialize").Return(nil)

	// Mock the Op method
	mockResponse := `
	<response status="success">
		<result>
			<devices>
				<entry>
					<hostname>test-fw</hostname>
					<serial>12345</serial>
					<ip-address>192.168.1.1</ip-address>
					<model>PA-3260</model>
					<family>3200</family>
					<sw-version>10.1.0</sw-version>
				</entry>
			</devices>
		</result>
	</response>`
	mockClient.On("Op", "<show><devices><connected/></devices></show>", "", nil, nil).Return([]byte(mockResponse), nil)

	// Test
	devices, err := dm.getDevicesFromPanorama()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, "test-fw", devices[0]["hostname"])
	assert.Equal(t, "12345", devices[0]["serial"])
	assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])
	assert.Equal(t, "PA-3260", devices[0]["model"])
	assert.Equal(t, "3200", devices[0]["family"])
	assert.Equal(t, "10.1.0", devices[0]["sw-version"])

	mockClient.AssertExpectations(t)
}

func TestFilterDevices(t *testing.T) {
	l := logger.New(0, false)
	devices := []map[string]string{
		{"hostname": "fw-1-a"},
		{"hostname": "fw-2-b"},
		{"hostname": "fw-3-c"},
		{"hostname": "other-fw"},
	}

	t.Run("No Filter", func(t *testing.T) {
		filtered := filterDevices(devices, []string{}, l)
		assert.Len(t, filtered, 4)
	})

	t.Run("Single Filter", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"fw-1"}, l)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "fw-1-a", filtered[0]["hostname"])
	})

	t.Run("Multiple Filters", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"fw-1", "fw-2"}, l)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "fw-1-a", filtered[0]["hostname"])
		assert.Equal(t, "fw-2-b", filtered[1]["hostname"])
	})

	t.Run("No Matches", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"nonexistent"}, l)
		assert.Len(t, filtered, 0)
	})
}
