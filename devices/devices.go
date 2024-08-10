// Package devices/devices.go
package devices

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"

	"github.com/PaloAltoNetworks/pango"
	"github.com/cdot65/pan-os-cdss-certificate/config"
	"github.com/cdot65/pan-os-cdss-certificate/logger"
)

// GetDeviceList retrieves a list of devices based on configuration and filters.

// This function fetches device information either from a local inventory file or
// from Panorama, depending on the noPanorama flag. It can also filter devices
// based on a provided hostname filter.

// Attributes:
//   conf (*config.Config): Configuration object for Panorama connection.
//   noPanorama (bool): Flag to determine whether to use local inventory or Panorama.
//   hostnameFilter (string): Comma-separated list of hostnames to filter devices.
//   l (*logger.Logger): Logger instance for logging operations.

// Error:
//   error: Returns an error if unable to read inventory file or fetch devices from Panorama.

// Return:
//   []map[string]string: List of devices, each represented as a map of string key-value pairs.
//   error: Any error encountered during the process.

func GetDeviceList(conf *config.Config, noPanorama bool, hostnameFilter string, l *logger.Logger) ([]map[string]string, error) {
	var deviceList []map[string]string
	var err error

	if noPanorama {
		inventory, err := readInventoryFile("inventory.yaml")
		if err != nil {
			return nil, fmt.Errorf("failed to read inventory file: %w", err)
		}
		deviceList = convertInventoryToDeviceList(inventory)
	} else {
		client := initializePanoramaClient(conf, l)
		deviceList, err = getConnectedDevices(client, l)
		if err != nil {
			return nil, fmt.Errorf("failed to get connected devices: %w", err)
		}
	}

	if hostnameFilter != "" {
		deviceList = filterDevices(deviceList, strings.Split(hostnameFilter, ","), l)
	}

	return deviceList, nil
}

// initializePanoramaClient initializes and returns a Panorama client.

// This function sets up a Panorama client using the provided configuration and logger.
// It initializes the client with the first Panorama configuration found in the config file.

// Attributes:
//   conf (*config.Config): Configuration object containing Panorama settings.
//   l (*logger.Logger): Logger object for logging operations.

// Error:
//   Fatal: If no Panorama configuration is found or client initialization fails.

// Return:
//   *pango.Panorama: An initialized Panorama client.

func initializePanoramaClient(conf *config.Config, l *logger.Logger) *pango.Panorama {
	if len(conf.Panorama) == 0 {
		l.Fatalf("No Panorama configuration found in the YAML file")
	}

	// Use the first Panorama configuration
	pano := conf.Panorama[0]

	// Initialize the Panorama client
	client := &pango.Panorama{
		Client: pango.Client{
			Hostname: pano.Hostname,
			Username: conf.Auth.Auth.Panorama.Username,
			Password: conf.Auth.Auth.Panorama.Password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}

	l.Info("Initializing client for", pano.Hostname)
	if err := client.Initialize(); err != nil {
		l.Fatalf("Failed to initialize client: %v", err)
	}
	l.Info("Client initialized for", pano.Hostname)

	return client
}

// getConnectedDevices retrieves a list of connected devices from a Panorama instance.

// This function sends a command to Panorama to fetch connected devices, parses the XML response,
// and returns a list of device information as key-value pairs.

// Attributes:
//   client (*pango.Panorama): Panorama client for API communication
//   l (*logger.Logger): Logger instance for debug output

// Error:
//   error: Returned if the operation fails due to API errors or response parsing issues

// Return:
//   []map[string]string: List of connected devices with their details
//   error: Any error encountered during the process

func getConnectedDevices(client *pango.Panorama, l *logger.Logger) ([]map[string]string, error) {
	cmd := "<show><devices><connected/></devices></show>"
	l.Debug("Sending command to get connected devices")
	response, err := client.Op(cmd, "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to perform op command: %w", err)
	}
	l.Debug("Received response for connected devices")

	var resp config.DevicesResponse
	if err := xml.Unmarshal(response, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("operation failed: %s", resp.Status)
	}

	var deviceList []map[string]string
	l.Debug("Number of devices found:", len(resp.Result.Devices.Entries))
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
		}
		deviceList = append(deviceList, device)
		l.Debug("Added device to list:", entry.Hostname)
	}

	l.Debug("Total devices in list:", len(deviceList))
	return deviceList, nil
}

// convertInventoryToDeviceList converts an Inventory struct to a list of device maps.

// This function takes an Inventory struct and transforms it into a slice of maps,
// where each map represents a device with its hostname and IP address.

// Attributes:
//   inventory (*config.Inventory): A pointer to the Inventory struct containing device information.

// Return:
//   deviceList ([]map[string]string): A slice of maps, each containing a device's hostname and IP address.

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
// containing only the devices whose hostnames match any of the given filters.
// It also logs debug and info messages about the filtering process.

// Attributes:
//   devices ([]map[string]string): List of device maps containing device information.
//   filters ([]string): List of hostname filters to apply.
//   l (*logger.Logger): Logger instance for logging debug and info messages.

// Return:
//   filteredDevices ([]map[string]string): List of devices that match the filters.

func filterDevices(devices []map[string]string, filters []string, l *logger.Logger) []map[string]string {
	if len(filters) == 0 {
		return devices
	}

	var filteredDevices []map[string]string
	for _, device := range devices {
		hostname := device["hostname"]
		for _, filter := range filters {
			if strings.Contains(hostname, strings.TrimSpace(filter)) {
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

// Attributes:
//   filename (string): The path to the inventory file to be read.

// Error:
//   error: File read error or YAML unmarshaling error.

// Return:
//   *config.Inventory: Pointer to the parsed Inventory struct.
//   error: Any error encountered during file reading or YAML parsing.

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
