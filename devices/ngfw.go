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

// defaultNgfwClientFactory is a function that creates a PAN-OS client for NGFW with the given hostname, username, and password.
// It returns a PanosClient interface that can be used for PAN-OS operations.
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

// getDevicesFromInventory retrieves the devices from the inventory file and
// collects their information by initializing the NGFW client for each device.
// It returns a list of devices as an array of maps, where each map contains
// the device information. If any errors occur during the retrieval process,
// an error is returned.
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
				dm.config.Auth.Credentials.Firewall.Username,
				dm.config.Auth.Credentials.Firewall.Password,
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

// getNgfwDeviceInfo retrieves the device information from a specific NGFW device using the provided PanosClient and hostname.
// It sends an "op" command to the device to get the system information.
// The method returns a map of device information, including serial number, hostname, IP address, model, software version,
// application version, antivirus version, Wildfire version, and threat version.
// If any errors occur during the process, an error is returned.
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
		"family":           resp.Result.System.Family,
		"sw-version":       resp.Result.System.SWVersion,
		"app-version":      resp.Result.System.AppVersion,
		"av-version":       resp.Result.System.AVVersion,
		"wildfire-version": resp.Result.System.WildfireVersion,
		"threat-version":   resp.Result.System.ThreatVersion,
		"result":           "",
	}, nil
}

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

// ConvertInventoryToDeviceList converts the devices in an inventory to a list of maps with "hostname" and "ip-address" keys.
// It takes in an `inventory` of type `*config.Inventory` and returns a slice of `map[string]string`.
// Each map in the slice represents a device in the inventory, with "hostname" as the key and the device's hostname as the value,
// and "ip-address" as the key and the device's IP address as the value.
// The function iterates over the devices in the inventory and appends a map for each device to the `deviceList` slice.
// Finally, it returns the `deviceList` slice.
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
