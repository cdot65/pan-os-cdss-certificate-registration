// Package devices devices/ngfw.go
package devices

import (
	"encoding/xml"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
	"sync"

	"github.com/PaloAltoNetworks/pango"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
)

// NGFWManager handles NGFW-specific operations
type NGFWManager struct {
	config *config.Config
	logger *logger.Logger
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

func (dm *DeviceManager) getDevicesFromInventory(inventory *config.Inventory) ([]map[string]string, error) {
	var deviceList []map[string]string
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(inventory.Inventory))

	for _, device := range inventory.Inventory {
		wg.Add(1)
		go func(device config.InventoryDevice) {
			defer wg.Done()

			ngfwClient := dm.panosClientFactory(
				device.IPAddress,
				dm.config.Auth.Auth.Firewall.Username,
				dm.config.Auth.Auth.Firewall.Password,
			)

			dm.logger.Info("Initializing NGFW client for", device.Hostname)
			if err := ngfwClient.Initialize(); err != nil {
				dm.logger.Debug(fmt.Sprintf("Failed to initialize NGFW client for %s: %v", device.Hostname, err))
				errChan <- err
				return
			}

			deviceInfo, err := dm.getDeviceInfo(ngfwClient, device.Hostname)
			if err != nil {
				dm.logger.Debug(fmt.Sprintf("Failed to get device info for %s: %v", device.Hostname, err))
				errChan <- err
				return
			}

			mu.Lock()
			deviceList = append(deviceList, deviceInfo)
			mu.Unlock()
		}(device)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return deviceList, <-errChan // Return partial results with the first error
	}

	return deviceList, nil
}

// GetDeviceList retrieves device information from NGFWs concurrently
func (nm *NGFWManager) GetDeviceList(inventory *config.Inventory) ([]config.DeviceEntry, error) {
	var wg sync.WaitGroup
	deviceChan := make(chan config.DeviceEntry, len(inventory.Inventory))
	errorChan := make(chan error, len(inventory.Inventory))

	for _, device := range inventory.Inventory {
		wg.Add(1)
		go func(dev config.InventoryDevice) {
			defer wg.Done()
			deviceInfo, err := nm.getDeviceInfo(dev)
			if err != nil {
				errorChan <- err
				return
			}
			deviceChan <- deviceInfo
		}(device)
	}

	go func() {
		wg.Wait()
		close(deviceChan)
		close(errorChan)
	}()

	var devices []config.DeviceEntry
	for device := range deviceChan {
		devices = append(devices, device)
	}

	if len(errorChan) > 0 {
		return nil, <-errorChan
	}

	return devices, nil
}

func (nm *NGFWManager) getDeviceInfo(device config.InventoryDevice) (config.DeviceEntry, error) {
	fw := &pango.Firewall{Client: pango.Client{
		Hostname: device.IPAddress,
		Username: nm.config.Auth.Auth.Firewall.Username,
		Password: nm.config.Auth.Auth.Firewall.Password,
		Logging:  pango.LogQuiet,
	}}

	if err := fw.Initialize(); err != nil {
		return config.DeviceEntry{}, fmt.Errorf("failed to initialize firewall connection: %v", err)
	}

	cmd := "<show><system><info/></system></show>"
	resp, err := fw.Op(cmd, "", nil, nil)
	if err != nil {
		return config.DeviceEntry{}, fmt.Errorf("failed to get system info: %v", err)
	}

	var systemInfo struct {
		XMLName xml.Name `xml:"response"`
		Status  string   `xml:"status,attr"`
		Result  struct {
			System config.DeviceEntry `xml:"system"`
		} `xml:"result"`
	}

	if err := xml.Unmarshal(resp, &systemInfo); err != nil {
		return config.DeviceEntry{}, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return systemInfo.Result.System, nil
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
