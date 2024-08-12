// Package devices/devices.go
package devices

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/pango"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
)

// PanoramaClient interface for the Panorama operations we need
type PanoramaClient interface {
	Initialize() error
	Op(cmd interface{}, vsys string, extras interface{}, ans interface{}) ([]byte, error)
}

// PanoramaClientFactory is a function type that creates a PanoramaClient
type PanoramaClientFactory func(hostname, username, password string) PanoramaClient

// DeviceManager handles device-related operations
type DeviceManager struct {
	config                *config.Config
	logger                *logger.Logger
	panoramaClient        PanoramaClient
	panoramaClientFactory PanoramaClientFactory
	inventoryReader       func(string) (*config.Inventory, error)
}

// NewDeviceManager creates a new DeviceManager
func NewDeviceManager(conf *config.Config, l *logger.Logger) *DeviceManager {
	return &DeviceManager{
		config:                conf,
		logger:                l,
		inventoryReader:       readInventoryFile,
		panoramaClientFactory: defaultPanoramaClientFactory,
	}
}

// defaultPanoramaClientFactory creates a real Panorama client
func defaultPanoramaClientFactory(hostname, username, password string) PanoramaClient {
	return &pango.Panorama{
		Client: pango.Client{
			Hostname: hostname,
			Username: username,
			Password: password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}
}

// SetPanoramaClientFactory sets a custom Panorama client factory
func (dm *DeviceManager) SetPanoramaClientFactory(factory PanoramaClientFactory) {
	dm.panoramaClientFactory = factory
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
		deviceList = convertInventoryToDeviceList(inventory)
	} else {
		if dm.panoramaClient == nil {
			dm.initializePanoramaClient()
		}
		deviceList, err = dm.getConnectedDevices()
		if err != nil {
			return nil, fmt.Errorf("failed to get connected devices: %w", err)
		}
	}

	if hostnameFilter != "" {
		deviceList = filterDevices(deviceList, strings.Split(hostnameFilter, ","), dm.logger)
	}

	return deviceList, nil
}

func (dm *DeviceManager) initializePanoramaClient() {
	if len(dm.config.Panorama) == 0 {
		dm.logger.Fatalf("No Panorama configuration found in the YAML file")
	}

	// Use the first Panorama configuration
	pano := dm.config.Panorama[0]

	dm.panoramaClient = dm.panoramaClientFactory(
		pano.Hostname,
		dm.config.Auth.Auth.Panorama.Username,
		dm.config.Auth.Auth.Panorama.Password,
	)

	dm.logger.Info("Initializing client for", pano.Hostname)
	if err := dm.panoramaClient.Initialize(); err != nil {
		dm.logger.Fatalf("Failed to initialize client: %v", err)
	}
	dm.logger.Info("Client initialized for", pano.Hostname)
}

func (dm *DeviceManager) getConnectedDevices() ([]map[string]string, error) {
	cmd := "<show><devices><connected/></devices></show>"
	dm.logger.Debug("Sending command to get connected devices")
	response, err := dm.panoramaClient.Op(cmd, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform op command: %w", err)
	}
	dm.logger.Debug("Received response for connected devices")

	var resp config.DevicesResponse
	if err := xml.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("operation failed: %s", resp.Status)
	}

	var deviceList []map[string]string
	dm.logger.Debug("Number of devices found:", len(resp.Result.Devices.Entries))
	for _, entry := range resp.Result.Devices.Entries {
		device := map[string]string{
			"serial":           entry.Serial,
			"hostname":         entry.Hostname,
			"ip-address":       entry.IPAddress,
			"ipv6-address":     entry.IPv6Address,
			"model":            entry.Model,
			"sw-version":       entry.SWVersion,
			"app-version":      entry.AppVersion,
			"av-version":       entry.AVVersion,
			"wildfire-version": entry.WildfireVersion,
			"threat-version":   entry.ThreatVersion,
			"result":           entry.Result,
		}
		deviceList = append(deviceList, device)
		dm.logger.Debug("Added device to list:", entry.Hostname)
	}

	dm.logger.Debug("Total devices in list:", len(deviceList))
	return deviceList, nil
}

// convertInventoryToDeviceList converts an Inventory struct to a list of device maps.
// This function takes an Inventory struct and transforms it into a slice of maps,
// where each map represents a device with its hostname and IP address.
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

// filterDevices filters a list of devices based on hostname filters.
// This function takes a list of devices and filters, and returns a new list
// containing only the devices whose hostnames start with any of the given filters.
// It also logs debug and info messages about the filtering process.
func filterDevices(devices []map[string]string, filters []string, l *logger.Logger) []map[string]string {
	if len(filters) == 0 {
		return devices
	}

	var filteredDevices []map[string]string
	for _, device := range devices {
		hostname := device["hostname"]
		for _, filter := range filters {
			if strings.HasPrefix(hostname, strings.TrimSpace(filter)) {
				filteredDevices = append(filteredDevices, device)
				l.Debug("Device matched filter:", hostname)
				break
			}
		}
	}

	l.Info("Filtered devices:", len(filteredDevices), "out of", len(devices))
	return filteredDevices
}

// readInventoryFile reads and parses an inventory file in YAML format.
// This function reads the contents of a file specified by the filename,
// and unmarshals the YAML data into a config.Inventory struct.
func readInventoryFile(filename string) (*config.Inventory, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var inventory config.Inventory
	err = yaml.Unmarshal(data, &inventory)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &inventory, nil
}
