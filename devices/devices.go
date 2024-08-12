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

// NewDeviceManager creates a new DeviceManager
func NewDeviceManager(conf *config.Config, l *logger.Logger) *DeviceManager {
	return &DeviceManager{
		config:             conf,
		logger:             l,
		inventoryReader:    readInventoryFile,
		panosClientFactory: defaultPanosClientFactory,
	}
}

// defaultPanosClientFactory creates a real PAN-OS client
func defaultPanosClientFactory(hostname, username, password string) PanosClient {
	return &pango.Firewall{
		Client: pango.Client{
			Hostname: hostname,
			Username: username,
			Password: password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}
}

// defaultPanoramaClientFactory creates a real Panorama client
func defaultPanoramaClientFactory(hostname, username, password string) PanosClient {
	return &pango.Panorama{
		Client: pango.Client{
			Hostname: hostname,
			Username: username,
			Password: password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}
}

// SetPanosClientFactory sets a custom PAN-OS client factory
func (dm *DeviceManager) SetPanosClientFactory(factory PanosClientFactory) {
	dm.panosClientFactory = factory
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

func (dm *DeviceManager) initializePanoramaClient() {
	if len(dm.config.Panorama) == 0 {
		dm.logger.Fatalf("No Panorama configuration found in the YAML file")
	}

	// Use the first Panorama configuration
	pano := dm.config.Panorama[0]

	dm.panosClient = defaultPanoramaClientFactory(
		pano.Hostname,
		dm.config.Auth.Auth.Panorama.Username,
		dm.config.Auth.Auth.Panorama.Password,
	)

	dm.logger.Info("Initializing Panorama client for", pano.Hostname)
	if err := dm.panosClient.Initialize(); err != nil {
		dm.logger.Fatalf("Failed to initialize Panorama client: %v", err)
	}
	dm.logger.Info("Panorama client initialized for", pano.Hostname)
}

func (dm *DeviceManager) getConnectedDevices() ([]map[string]string, error) {
	cmd := "<show><devices><connected/></devices></show>"
	dm.logger.Debug("Sending command to get connected devices")
	response, err := dm.panosClient.Op(cmd, "", nil, nil)
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

func (dm *DeviceManager) getDevicesFromInventory(inventory *config.Inventory) ([]map[string]string, error) {
	var deviceList []map[string]string
	for _, device := range inventory.Inventory {
		ngfwClient := dm.panosClientFactory(
			device.IPAddress,
			dm.config.Auth.Auth.Firewall.Username,
			dm.config.Auth.Auth.Firewall.Password,
		)

		dm.logger.Info("Initializing NGFW client for", device.Hostname)
		if err := ngfwClient.Initialize(); err != nil {
			dm.logger.Debug(fmt.Sprintf("Failed to initialize NGFW client for %s: %v", device.Hostname, err))
			continue
		}

		deviceInfo, err := dm.getDeviceInfo(ngfwClient, device.Hostname)
		if err != nil {
			dm.logger.Debug(fmt.Sprintf("Failed to get device info for %s: %v", device.Hostname, err))
			continue
		}

		deviceList = append(deviceList, deviceInfo)
	}

	return deviceList, nil
}

func (dm *DeviceManager) getDeviceInfo(client PanosClient, hostname string) (map[string]string, error) {
	cmd := "<show><system><info/></system></show>"
	response, err := client.Op(cmd, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform op command: %w %s", err, hostname)
	}

	var resp struct {
		XMLName xml.Name `xml:"response"`
		Status  string   `xml:"status,attr"`
		Result  struct {
			System config.DeviceEntry `xml:"system"`
		} `xml:"result"`
	}

	if err := xml.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("operation failed: %s", resp.Status)
	}

	return map[string]string{
		"serial":           resp.Result.System.Serial,
		"hostname":         resp.Result.System.Hostname,
		"ip-address":       resp.Result.System.IPAddress,
		"ipv6-address":     resp.Result.System.IPv6Address,
		"model":            resp.Result.System.Model,
		"sw-version":       resp.Result.System.SWVersion,
		"app-version":      resp.Result.System.AppVersion,
		"av-version":       resp.Result.System.AVVersion,
		"wildfire-version": resp.Result.System.WildfireVersion,
		"threat-version":   resp.Result.System.ThreatVersion,
		"result":           "",
	}, nil
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
