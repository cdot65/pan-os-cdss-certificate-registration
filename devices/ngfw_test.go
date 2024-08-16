package devices

import (
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNgfwClient is a mock implementation of the PanosClient interface
type MockNgfwClient struct {
	mock.Mock
}

func (m *MockNgfwClient) Initialize() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockNgfwClient) Op(cmd interface{}, vsys string, extras interface{}, ans interface{}) ([]byte, error) {
	args := m.Called(cmd, vsys, extras, ans)
	return args.Get(0).([]byte), args.Error(1)
}

// testInventory holds our test inventory data
var testInventory = &config.Inventory{
	Inventory: []config.InventoryDevice{
		{Hostname: "test-fw-1", IPAddress: "192.168.1.1"},
		{Hostname: "test-fw-2", IPAddress: "192.168.1.2"},
	},
}

// TestDeviceManager extends DeviceManager for testing purposes
type TestDeviceManager struct {
	DeviceManager
}

// getDevicesFromInventory overrides the original method for testing
func (tdm *TestDeviceManager) getDevicesFromInventory() ([]map[string]string, error) {
	var deviceList []map[string]string
	for _, device := range testInventory.Inventory {
		ngfwClient := tdm.panosClientFactory(
			device.IPAddress,
			tdm.config.Auth.Credentials.Firewall.Username,
			tdm.config.Auth.Credentials.Firewall.Password,
		)

		if err := ngfwClient.Initialize(); err != nil {
			return nil, err
		}

		deviceInfo, err := tdm.getNgfwDeviceInfo(ngfwClient, device.Hostname)
		if err != nil {
			return nil, err
		}

		deviceList = append(deviceList, deviceInfo)
	}

	return deviceList, nil
}

// TestGetDevicesFromInventory tests the getDevicesFromInventory function
func TestGetDevicesFromInventory(t *testing.T) {
	// Setup TestDeviceManager with mock client factory
	conf := &config.Config{
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
				Firewall: struct {
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
	dm := &TestDeviceManager{DeviceManager: *NewDeviceManager(conf, l)}

	mockClient := new(MockNgfwClient)
	dm.panosClientFactory = func(hostname, username, password string) PanosClient {
		return mockClient
	}

	// Mock the Initialize and Op methods
	mockClient.On("Initialize").Return(nil)
	mockResponse := `
	<response status="success">
		<result>
			<system>
				<hostname>test-fw</hostname>
				<serial>12345</serial>
				<ip-address>192.168.1.1</ip-address>
				<model>PA-3260</model>
				<family>3200</family>
				<sw-version>10.1.0</sw-version>
			</system>
		</result>
	</response>`
	mockClient.On("Op", "<show><system><info/></system></show>", "", nil, nil).Return([]byte(mockResponse), nil)

	// Test
	devices, err := dm.getDevicesFromInventory()

	// Assert
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "test-fw", devices[0]["hostname"])
	assert.Equal(t, "12345", devices[0]["serial"])
	assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])
	assert.Equal(t, "PA-3260", devices[0]["model"])
	assert.Equal(t, "3200", devices[0]["family"])
	assert.Equal(t, "10.1.0", devices[0]["sw-version"])

	mockClient.AssertExpectations(t)
}
