// devices/devices_test.go
package devices

import (
	"os"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/stretchr/testify/assert"
)

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
	assert.NoError(t, err)
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {
			assert.Error(t, err)
		}
	}(tmpfile.Name())

	_, err = tmpfile.Write([]byte(content))
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

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

// Note: We can't easily test GetDeviceList, initializePanoramaClient, and getConnectedDevices
// without mocking the Panorama client, which would require significant changes to the code.
// Ain't got time for that right now.
