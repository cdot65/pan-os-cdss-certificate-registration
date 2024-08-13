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
)

// defaultNgfwClientFactory creates a real PAN-OS client for NGFW
func defaultNgfwClientFactory(hostname, username, password string) PanosClient {
	return &pango.Firewall{
		Client: pango.Client{
			Hostname: hostname,
			Username: username,
			Password: password,
			Logging:  pango.LogAction | pango.LogOp,
		},
	}
}

func (dm *DeviceManager) getDevicesFromInventory() ([]map[string]string, error) {
	inventory, err := readInventoryFile("inventory.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read inventory file: %w", err)
	}

	var deviceList []map[string]string
	var mu sync.Mutex
	var wg sync.WaitGroup
	errorList := make([]string, 0)

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
				errorMsg := fmt.Sprintf("Failed to initialize NGFW client for %s: %v", device.Hostname, err)
				dm.logger.Debug(errorMsg)
				mu.Lock()
				errorList = append(errorList, errorMsg)
				mu.Unlock()
				return
			}

			deviceInfo, err := dm.getNgfwDeviceInfo(ngfwClient, device.Hostname)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get device info for %s: %v", device.Hostname, err)
				dm.logger.Debug(errorMsg)
				mu.Lock()
				errorList = append(errorList, errorMsg)
				mu.Unlock()
				return
			}

			mu.Lock()
			deviceList = append(deviceList, deviceInfo)
			mu.Unlock()
		}(device)
	}

	wg.Wait()

	// Print errors if any
	if len(errorList) > 0 {
		fmt.Println("Errors occurred while processing devices:")
		for _, errMsg := range errorList {
			fmt.Println(errMsg)
		}
		fmt.Println() // Add a blank line for better readability
	}

	return deviceList, nil
}

// getNgfwDeviceInfo retrieves device information from a single NGFW
func (dm *DeviceManager) getNgfwDeviceInfo(client PanosClient, hostname string) (map[string]string, error) {
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

// ConvertInventoryToDeviceList is used for testing purposes
func ConvertInventoryToDeviceList(inventory *config.Inventory) []map[string]string {
	var deviceList []map[string]string
	for _, device := range inventory.Inventory {
		deviceList = append(deviceList, map[string]string{
			"hostname":   device.Hostname,
			"ip-address": device.IPAddress,
		})
	}
	return deviceList
}
