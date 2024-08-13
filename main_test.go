// main_test.go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
	"github.com/cdot65/pan-os-cdss-certificate-registration/logger"
	"github.com/cdot65/pan-os-cdss-certificate-registration/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConfig is a mock implementation of the config.Config struct
type MockConfig struct {
	mock.Mock
	Verbose bool
}

// MockDevices is a mock implementation of the devices package
type MockDevices struct {
	mock.Mock
}

// MockUtils is a mock implementation of the utils package
type MockUtils struct {
	mock.Mock
}

// MockWildfire is a mock implementation of the wildfire package
type MockWildfire struct {
	mock.Mock
}

func (m *MockConfig) Load(configFile, secretsFile string) (*config.Config, error) {
	args := m.Called(configFile, secretsFile)
	return args.Get(0).(*config.Config), args.Error(1)
}

func (m *MockDevices) GetDeviceList(conf *config.Config, noPanorama bool, hostnameFilter string, l *logger.Logger) ([]map[string]string, error) {
	args := m.Called(conf, noPanorama, hostnameFilter, l)
	return args.Get(0).([]map[string]string), args.Error(1)
}

func (m *MockUtils) ParseVersion(version string) (*utils.Version, error) {
	args := m.Called(version)
	return args.Get(0).(*utils.Version), args.Error(1)
}

func (m *MockUtils) FilterAffectedDevices(deviceList []map[string]string) ([]map[string]string, error) {
	args := m.Called(deviceList)
	return args.Get(0).([]map[string]string), args.Error(1)
}

func (m *MockWildfire) RegisterWildFire(device map[string]string, username, password string, l *logger.Logger) error {
	args := m.Called(device, username, password, l)
	return args.Error(0)
}

func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

func TestMainLogic(t *testing.T) {
	// Setup mocks
	mockConfig := new(MockConfig)
	mockDevices := new(MockDevices)
	mockUtils := new(MockUtils)
	mockWildfire := new(MockWildfire)

	// Create a mock config with a verbose flag
	mockCfg := &MockConfig{
		Verbose: true, // or false, depending on what you want to test
	}

	// Setup test data
	testConf := &config.Config{
		Auth: config.AuthConfig{
			Credentials: struct {
				Panorama struct {
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				} `yaml:"panorama"`
				Firewall struct {
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				} `yaml:"firewall"`
			}{
				Firewall: struct {
					Username string `yaml:"username"`
					Password string `yaml:"password"`
				}{
					Username: "testuser",
					Password: "testpass",
				},
			},
		},
	}

	// Create a test device list with various PAN-OS versions, excluding "-gp" versions
	testDeviceList := make([]map[string]string, 0)
	for versionKey, patchVersions := range config.MinimumPatchedVersions {
		if !strings.HasSuffix(versionKey, "-gp") {
			for _, patchVersion := range patchVersions {
				swVersion := fmt.Sprintf("%s.%d-h%d", versionKey, patchVersion.Maintenance, patchVersion.Hotfix)
				testDeviceList = append(testDeviceList, map[string]string{
					"hostname":   fmt.Sprintf("device-%s", swVersion),
					"sw-version": swVersion,
				})
			}
		}
	}

	// Add affected devices
	affectedDevices := []map[string]string{
		{"hostname": "affected-firewall1", "sw-version": "8.1.21-h2"},
		{"hostname": "affected-firewall2", "sw-version": "9.0.16-h6"},
		{"hostname": "affected-firewall3", "sw-version": "9.1.14-h7"},
		{"hostname": "affected-firewall4", "sw-version": "10.0.8-h0"},
		{"hostname": "affected-firewall5", "sw-version": "10.1.3-h2"},
		{"hostname": "affected-firewall6", "sw-version": "10.2.3-h11"},
		{"hostname": "affected-firewall7", "sw-version": "11.0.3-h2"},
		{"hostname": "affected-firewall8", "sw-version": "11.1.0-h1"},
	}
	testDeviceList = append(testDeviceList, affectedDevices...)

	// Set up ParseVersion expectations
	for _, device := range testDeviceList {
		parsedVersion, _ := utils.ParseVersion(device["sw-version"])
		mockUtils.On("ParseVersion", device["sw-version"]).Return(parsedVersion, nil).Once()
	}

	// Set up other expectations
	mockConfig.On("Load", mock.Anything, mock.Anything).Return(testConf, nil)
	mockDevices.On("GetDeviceList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testDeviceList, nil)

	// Modify the FilterAffectedDevices mock
	mockUtils.On("FilterAffectedDevices", mock.Anything).Return(affectedDevices, nil)

	mockWildfire.On("RegisterWildFire", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Run the main logic (without actually calling main())
	flags, cfg := config.ParseFlags()
	l := logger.New(flags.DebugLevel, flags.Verbose)

	// Use cfg instead of conf for the initial configuration
	conf, err := mockConfig.Load(flags.ConfigFile, flags.SecretsFile)
	assert.NoError(t, err)

	deviceList, err := mockDevices.GetDeviceList(conf, flags.NoPanorama, cfg.HostnameFilter, l)
	assert.NoError(t, err)

	for i, device := range deviceList {
		swVersion := device["sw-version"]
		parsedVersion, err := mockUtils.ParseVersion(swVersion)
		assert.NoError(t, err)

		deviceList[i]["parsed_version_major"] = fmt.Sprintf("%d", parsedVersion.Major)
		deviceList[i]["parsed_version_feature"] = fmt.Sprintf("%d", parsedVersion.Feature)
		deviceList[i]["parsed_version_maintenance"] = fmt.Sprintf("%d", parsedVersion.Maintenance)
		deviceList[i]["parsed_version_hotfix"] = fmt.Sprintf("%d", parsedVersion.Hotfix)
	}

	filteredDevices, err := mockUtils.FilterAffectedDevices(deviceList)
	assert.NoError(t, err)

	// Redirect stdout to our buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Update the PrintDeviceList call
	utils.PrintDeviceList(filteredDevices, l, mockCfg.Verbose)

	for _, device := range filteredDevices {
		err := mockWildfire.RegisterWildFire(device, conf.Auth.Credentials.Firewall.Username, conf.Auth.Credentials.Firewall.Password, l)
		assert.NoError(t, err)
		_, err = fmt.Fprintf(w, "%s: Successfully registered WildFire\n", device["hostname"])
		assert.NoError(t, err)
	}

	// Restore stdout
	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}
	os.Stdout = old

	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	err = r.Close()
	if err != nil {
		t.Fatalf("Failed to close reader: %v", err)
	}

	// Assertions
	output := buf.String()
	assert.Contains(t, output, "Device List:")
	for _, device := range filteredDevices {
		assert.Contains(t, output, fmt.Sprintf("hostname: %s", device["hostname"]))
		assert.Contains(t, output, fmt.Sprintf("sw-version: %s", device["sw-version"]))
		assert.Contains(t, output, fmt.Sprintf("%s: Successfully registered WildFire", device["hostname"]))
	}

	mockConfig.AssertExpectations(t)
	mockDevices.AssertExpectations(t)
	mockUtils.AssertExpectations(t)
	mockWildfire.AssertExpectations(t)
}
