package devices

import (
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

func TestConvertInventoryToDeviceList(t *testing.T) {
	inventory := &config.Inventory{
		Inventory: []config.InventoryDevice{
			{Hostname: "device1", IPAddress: "192.168.1.1"},
			{Hostname: "device2", IPAddress: "192.168.1.2"},
		},
	}

	deviceList := convertInventoryToDeviceList(inventory)

	assert.Len(t, deviceList, 2)
	assert.Equal(t, "device1", deviceList[0]["hostname"])
	assert.Equal(t, "192.168.1.1", deviceList[0]["ip-address"])
	assert.Equal(t, "device2", deviceList[1]["hostname"])
	assert.Equal(t, "192.168.1.2", deviceList[1]["ip-address"])
}

func TestFilterDevices(t *testing.T) {
	devices := []map[string]string{
		{"hostname": "device1", "ip-address": "192.168.1.1"},
		{"hostname": "device2", "ip-address": "192.168.1.2"},
		{"hostname": "other-device", "ip-address": "192.168.1.3"},
	}

	l := logger.New(0, false)

	t.Run("Filter single device", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"device1"}, l)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "device1", filtered[0]["hostname"])
	})

	t.Run("Filter multiple devices", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"device"}, l)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "device1", filtered[0]["hostname"])
		assert.Equal(t, "device2", filtered[1]["hostname"])
	})

	t.Run("No matching filter", func(t *testing.T) {
		filtered := filterDevices(devices, []string{"nonexistent"}, l)
		assert.Len(t, filtered, 0)
	})

	t.Run("Empty filter", func(t *testing.T) {
		filtered := filterDevices(devices, []string{}, l)
		assert.Len(t, filtered, 3)
	})
}

func TestGetDevicesFromInventory(t *testing.T) {
	conf := &config.Config{}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	inventory := &config.Inventory{
		Inventory: []config.InventoryDevice{
			{Hostname: "device1", IPAddress: "192.168.1.1"},
			{Hostname: "device2", IPAddress: "192.168.1.2"},
		},
	}

	mockClient := new(MockPanosClient)
	mockClient.On("Initialize").Return(nil)
	mockClient.On("Op", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(`
		<response status="success">
			<result>
				<system>
					<hostname>device1</hostname>
					<ip-address>192.168.1.1</ip-address>
					<model>PA-220</model>
					<sw-version>10.1.0</sw-version>
				</system>
			</result>
		</response>
	`), nil)

	dm.SetPanosClientFactory(func(hostname, username, password string) PanosClient {
		return mockClient
	})

	devices, err := dm.getDevicesFromInventory(inventory)
	assert.NoError(t, err)
	assert.Len(t, devices, 2)
	assert.Equal(t, "device1", devices[0]["hostname"])
	assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])

	mockClient.AssertExpectations(t)
}

func TestGetConnectedDevices(t *testing.T) {
	conf := &config.Config{}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	mockClient := new(MockPanosClient)
	mockClient.On("Op", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte(`
		<response status="success">
			<result>
				<devices>
					<entry>
						<hostname>firewall1</hostname>
						<serial>1234</serial>
						<ip-address>192.168.1.1</ip-address>
						<sw-version>10.1.0</sw-version>
					</entry>
				</devices>
			</result>
		</response>
	`), nil)

	dm.panosClient = mockClient

	devices, err := dm.getConnectedDevices()
	assert.NoError(t, err)
	assert.Len(t, devices, 1)
	assert.Equal(t, "firewall1", devices[0]["hostname"])
	assert.Equal(t, "1234", devices[0]["serial"])
	assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])
	assert.Equal(t, "10.1.0", devices[0]["sw-version"])

	mockClient.AssertExpectations(t)
}

func TestInitializePanoramaClient(t *testing.T) {
	conf := &config.Config{
		Panorama: []struct {
			Hostname string `yaml:"hostname"`
		}{{Hostname: "panorama.example.com"}},
		Auth: config.AuthConfig{
			Auth: struct {
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
					Username: "admin",
					Password: "password",
				},
			},
		},
	}
	l := logger.New(0, false)
	dm := NewDeviceManager(conf, l)

	mockClient := new(MockPanosClient)
	mockClient.On("Initialize").Return(nil)

	// Replace the defaultPanoramaClientFactory with a mock factory
	oldFactory := defaultPanoramaClientFactory
	defaultPanoramaClientFactory = func(hostname, username, password string) PanosClient {
		assert.Equal(t, "panorama.example.com", hostname)
		assert.Equal(t, "admin", username)
		assert.Equal(t, "password", password)
		return mockClient
	}
	defer func() { defaultPanoramaClientFactory = oldFactory }()

	dm.initializePanoramaClient()

	mockClient.AssertExpectations(t)
	assert.Equal(t, mockClient, dm.panosClient)
}
