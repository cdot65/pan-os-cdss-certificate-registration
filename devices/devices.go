// Package devices devices/devices.go
package devices

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"strings"
)

// PanosClient interface for the PAN-OS operations we need
type PanosClient interface {
	Initialize() error
	Op(cmd interface{}, vsys string, extras interface{}, ans interface{}) ([]byte, error)
}

// PanosClientFactory is a function type that creates a PanosClient
type PanosClientFactory func(hostname, username, password string) PanosClient

// DeviceManager handles device-related operations
type DeviceManager struct {
	config             *config.Config
	logger             *logger.Logger
	panosClient        PanosClient
	panosClientFactory PanosClientFactory
	inventoryReader    func(string) (*config.Inventory, error)
}

// Used in tests only
func convertInventoryToDeviceList(inventory *config.Inventory) []map[string]string {
	var deviceList []map[string]string
	for _, device := range inventory.Inventory {
		deviceList = append(deviceList, map[string]string{
			"hostname":   device.Hostname,
			"ip-address": device.IPAddress,
		})
	}
	return deviceList
}

// NewDeviceManager creates a new DeviceManager
func NewDeviceManager(conf *config.Config, l *logger.Logger) *DeviceManager {
	return &DeviceManager{
		config:             conf,
		logger:             l,
		inventoryReader:    readInventoryFile,
		panosClientFactory: defaultPanosClientFactory,
	}
}

// GetDeviceList retrieves a list of devices based on configuration and filters.
func (dm *DeviceManager) GetDeviceList(noPanorama bool, hostnameFilter string) ([]map[string]string, error) {
	var deviceList []map[string]string
	var err error

	if noPanorama {
		inventory, err := dm.inventoryReader("inventory.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read inventory file: %w", err)
		}
		deviceList, err = dm.getDevicesFromInventory(inventory)
	} else {
		if dm.panosClient == nil {
			dm.initializePanoramaClient()
		}
		deviceList, err = dm.getConnectedDevices()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if hostnameFilter != "" {
		deviceList = filterDevices(deviceList, strings.Split(hostnameFilter, ","), dm.logger)
	}

	return deviceList, nil
}

// SetPanosClientFactory sets a custom PAN-OS client factory
func (dm *DeviceManager) SetPanosClientFactory(factory PanosClientFactory) {
	dm.panosClientFactory = factory
}
