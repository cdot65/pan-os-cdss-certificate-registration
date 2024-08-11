// devices/devices_test.go
package devices

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPanoramaClient is a mock implementation of the PanoramaClient interface
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

func TestConvertInventoryToDeviceList(t *testing.T) {
	inventory := &config.Inventory{
		Inventory: []struct {
			Hostname  string `yaml:"hostname"`
			IPAddress string `yaml:"ip_address"`
		}{
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

func TestReadInventoryFile(t *testing.T) {
	// Create a temporary inventory file
	content := `
inventory:
  - hostname: device1
    ip_address: 192.168.1.1
  - hostname: device2
    ip_address: 192.168.1.2
`
	tmpfile, err := os.CreateTemp("", "inventory*.yaml")
	require.NoError(t, err)
	defer func() {
		err := os.Remove(tmpfile.Name())
		assert.NoError(t, err, "Failed to remove temporary file")
	}()

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err)
	err = tmpfile.Close()
	require.NoError(t, err)

	// Test reading the inventory file
	inventory, err := readInventoryFile(tmpfile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, inventory)
	assert.Len(t, inventory.Inventory, 2)
	assert.Equal(t, "device1", inventory.Inventory[0].Hostname)
	assert.Equal(t, "192.168.1.1", inventory.Inventory[0].IPAddress)
	assert.Equal(t, "device2", inventory.Inventory[1].Hostname)
	assert.Equal(t, "192.168.1.2", inventory.Inventory[1].IPAddress)
}

func TestReadInventoryFileError(t *testing.T) {
	_, err := readInventoryFile("nonexistent_file.yaml")
	assert.Error(t, err)
}

func TestGetDeviceList(t *testing.T) {
	conf := &config.Config{
		Panorama: []struct {
			Hostname string `yaml:"hostname"`
		}([]struct{ Hostname string }{
			{Hostname: "panorama.example.com"},
		}),
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

	t.Run("Get devices from Panorama", func(t *testing.T) {
		mockClient := new(MockPanoramaClient)
		mockClient.On("Initialize").Return(nil)
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

		dm.SetPanoramaClientFactory(func(hostname, username, password string) PanoramaClient {
			return mockClient
		})

		devices, err := dm.GetDeviceList(false, "")
		assert.NoError(t, err)
		assert.Len(t, devices, 1)
		assert.Equal(t, "firewall1", devices[0]["hostname"])
		assert.Equal(t, "1234", devices[0]["serial"])
		assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])
		assert.Equal(t, "10.1.0", devices[0]["sw-version"])

		mockClient.AssertExpectations(t)
	})

	t.Run("Get devices from inventory file", func(t *testing.T) {
		dm.inventoryReader = func(filename string) (*config.Inventory, error) {
			return &config.Inventory{
				Inventory: []struct {
					Hostname  string `yaml:"hostname"`
					IPAddress string `yaml:"ip_address"`
				}{
					{Hostname: "device1", IPAddress: "192.168.1.1"},
					{Hostname: "device2", IPAddress: "192.168.1.2"},
				},
			}, nil
		}

		devices, err := dm.GetDeviceList(true, "")
		assert.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Equal(t, "device1", devices[0]["hostname"])
		assert.Equal(t, "192.168.1.1", devices[0]["ip-address"])
		assert.Equal(t, "device2", devices[1]["hostname"])
		assert.Equal(t, "192.168.1.2", devices[1]["ip-address"])
	})

	t.Run("Filter devices", func(t *testing.T) {
		dm.inventoryReader = func(filename string) (*config.Inventory, error) {
			return &config.Inventory{
				Inventory: []struct {
					Hostname  string `yaml:"hostname"`
					IPAddress string `yaml:"ip_address"`
				}{
					{Hostname: "device1", IPAddress: "192.168.1.1"},
					{Hostname: "device2", IPAddress: "192.168.1.2"},
					{Hostname: "other-device", IPAddress: "192.168.1.3"},
				},
			}, nil
		}

		devices, err := dm.GetDeviceList(true, "device")
		assert.NoError(t, err)
		assert.Len(t, devices, 2)
		assert.Equal(t, "device1", devices[0]["hostname"])
		assert.Equal(t, "device2", devices[1]["hostname"])
	})
}

func TestInitializePanoramaClient(t *testing.T) {
	conf := &config.Config{
		Panorama: []struct {
			Hostname string `yaml:"hostname"`
		}([]struct{ Hostname string }{
			{Hostname: "panorama.example.com"},
		}),
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

	mockClient := new(MockPanoramaClient)
	mockClient.On("Initialize").Return(nil)

	dm.SetPanoramaClientFactory(func(hostname, username, password string) PanoramaClient {
		assert.Equal(t, "panorama.example.com", hostname)
		assert.Equal(t, "admin", username)
		assert.Equal(t, "password", password)
		return mockClient
	})

	dm.initializePanoramaClient()

	mockClient.AssertExpectations(t)
}
