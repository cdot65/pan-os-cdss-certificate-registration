// Package devices devices/devices.go
package devices

import (
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
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
}

// NewDeviceManager creates a new instance of DeviceManager with the provided configuration and logger.
// The panosClientFactory field is set to nil initially and can be set later based on the workflow.
// The function returns a pointer to the created DeviceManager.
func NewDeviceManager(conf *config.Config, l *logger.Logger) *DeviceManager {
	return &DeviceManager{
		config:             conf,
		logger:             l,
		panosClientFactory: nil, // This is set this later based on the workflow
	}
}

// GetDeviceList retrieves a list of devices and their information.
// If noPanorama is true, it retrieves the devices from the local inventory file.
// If noPanorama is false, it retrieves the devices from Panorama.
// The hostnameFilter parameter can be used to filter the devices based on their hostname.
// It returns the list of devices as an array of maps, where each map contains the device information.
func (dm *DeviceManager) GetDeviceList(noPanorama bool, hostnameFilter string) ([]map[string]string, error) {
	if dm.panosClientFactory == nil {
		if noPanorama {
			dm.SetNgfwWorkflow()
		} else {
			dm.SetPanoramaWorkflow()
		}
	}

	var deviceList []map[string]string
	var err error

	if noPanorama {
		deviceList, err = dm.getDevicesFromInventory()
	} else {
		if dm.panosClient == nil {
			dm.initializePanoramaClient()
		}
		deviceList, err = dm.getDevicesFromPanorama()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return deviceList, nil
}

// SetPanosClientFactory sets a custom PAN-OS client factory
func (dm *DeviceManager) SetPanosClientFactory(factory PanosClientFactory) {
	dm.panosClientFactory = factory
}

// SetNgfwWorkflow sets the PAN-OS client factory to create a real PAN-OS client for NGFW.
func (dm *DeviceManager) SetNgfwWorkflow() {
	dm.panosClientFactory = defaultNgfwClientFactory
}

// SetPanoramaWorkflow sets the PAN-OS client factory to create a real Panorama client.
func (dm *DeviceManager) SetPanoramaWorkflow() {
	dm.panosClientFactory = defaultPanoramaClientFactory
}
