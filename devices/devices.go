// Package devices devices/devices.go
package devices

import (
	"encoding/json"
	"fmt"
	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"sync"
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
	panosClientFactory PanosClientFactory
}

// NewDeviceManager creates a new instance of DeviceManager with the provided configuration and logger.
// The panosClientFactory field is set to nil initially and can be set later based on the workflow.
// The function returns a pointer to the created DeviceManager.
func NewDeviceManager(conf *config.Config, l *logger.Logger) *DeviceManager {
	return &DeviceManager{
		config:             conf,
		logger:             l,
		panosClientFactory: nil, // This is set later based on the workflow
	}
}

// GetDeviceList retrieves a list of devices and their information.
// If noPanorama is true, it retrieves the devices from the local inventory file.
// If noPanorama is false, it retrieves the devices from Panorama.
// It returns the list of devices as an array of maps, where each map contains the device information.
func (dm *DeviceManager) GetDeviceList(noPanorama bool) ([]map[string]string, error) {
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
		deviceList, err = dm.getDevicesFromPanorama()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	return deviceList, nil
}

// GetDeviceCertificateStatus retrieves the output from the command `show device-certificate status`
// It will always leverage the pango SDK, and only interact with NGFW devices
// It will update each device in the deviceList with the certificate status information
func (dm *DeviceManager) GetDeviceCertificateStatus(deviceList []map[string]string) {
	// Always set to NGFW workflow for this operation
	dm.SetNgfwWorkflow()

	var wg sync.WaitGroup

	for i := range deviceList {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			device := deviceList[index]
			hostname := device["hostname"]
			ipAddress := device["ip-address"]

			// Initialize the errors slice if it doesn't exist
			if _, ok := device["errors"]; !ok {
				deviceList[index]["errors"] = "[]"
			}

			// Create a new pango client for each device
			client := dm.panosClientFactory(
				ipAddress,
				dm.config.Auth.Credentials.Firewall.Username,
				dm.config.Auth.Credentials.Firewall.Password,
			)

			// Initialize the client
			if err := client.Initialize(); err != nil {
				errMsg := fmt.Sprintf("Failed to initialize client for %s: %v", hostname, err)
				dm.logger.Error(errMsg)
				deviceList[index]["errors"] = appendError(deviceList[index]["errors"], errMsg)
				return
			}

			// Get device certificate status
			certStatus, err := dm.showDeviceCertificateStatus(client, hostname)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to get device certificate status for %s: %v", hostname, err)
				dm.logger.Error(errMsg)
				deviceList[index]["errors"] = appendError(deviceList[index]["errors"], errMsg)
				return
			}

			// Update the device entry with certificate status information
			deviceList[index]["deviceCert"] = certStatusToJSON(certStatus)
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Log a summary of errors
	errorCount := 0
	for _, device := range deviceList {
		if errors := device["errors"]; errors != "" {
			errorCount++
			dm.logger.Warn(fmt.Sprintf("Device %s encountered errors: %s", device["hostname"], errors))
		}
	}
	if errorCount > 0 {
		dm.logger.Warn(fmt.Sprintf("Total devices encountered errors while getting device certificate status: %d", errorCount))
	} else {
		dm.logger.Info("Successfully retrieved device certificate status for all devices")
	}
}

// SetNgfwWorkflow sets the PAN-OS client factory to create a real PAN-OS client for NGFW.
func (dm *DeviceManager) SetNgfwWorkflow() {
	dm.panosClientFactory = defaultNgfwClientFactory
}

// SetPanoramaWorkflow sets the PAN-OS client factory to create a real Panorama client.
func (dm *DeviceManager) SetPanoramaWorkflow() {
	dm.panosClientFactory = defaultPanoramaClientFactory
}

func certStatusToJSON(certStatus map[string]string) string {
	jsonBytes, err := json.Marshal(certStatus)
	if err != nil {
		return "{}"
	}
	return string(jsonBytes)
}

func appendError(errorsJSON, newError string) string {
	var errors []string
	if err := json.Unmarshal([]byte(errorsJSON), &errors); err != nil {
		// If we can't unmarshal, start with an empty slice
		errors = []string{}
	}
	errors = append(errors, newError)
	jsonBytes, err := json.Marshal(errors)
	if err != nil {
		// If we can't marshal, return a JSON array with just the new error
		return fmt.Sprintf("[%q]", newError)
	}
	return string(jsonBytes)
}
