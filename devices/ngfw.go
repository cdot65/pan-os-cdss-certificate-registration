package devices

import (
	"encoding/xml"
	"fmt"
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
