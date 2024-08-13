// Package devices devices/panorama.go
package devices

import (
	"encoding/xml"
	"fmt"
	"github.com/PaloAltoNetworks/pango"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"strings"
)

// defaultPanoramaClientFactory creates a real Panorama client
var defaultPanoramaClientFactory = func(hostname, username, password string) PanosClient {
	return &pango.Panorama{
		Client: pango.Client{
			Hostname: hostname,
			Username: username,
			Password: password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}
}

func (dm *DeviceManager) initializePanoramaClient() {
	if len(dm.config.Panorama) == 0 {
		dm.logger.Fatalf("No Panorama configuration found in the YAML file")
	}

	// Use the first Panorama configuration
	pano := dm.config.Panorama[0]

	dm.panosClient = defaultPanoramaClientFactory(
		pano.Hostname,
		dm.config.Auth.Credentials.Panorama.Username,
		dm.config.Auth.Credentials.Panorama.Password,
	)

	dm.logger.Info("Initializing Panorama client for", pano.Hostname)
	if err := dm.panosClient.Initialize(); err != nil {
		dm.logger.Fatalf("Failed to initialize Panorama client: %v", err)
	}
	dm.logger.Info("Panorama client initialized for", pano.Hostname)
}

func (dm *DeviceManager) getDevicesFromPanorama() ([]map[string]string, error) {
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
