package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cdot65/pan-os-cdss-certificate-registration/config"
)

// Version represents a PAN-OS version
type Version struct {
	Major       int
	Feature     int
	Maintenance int
	Hotfix      int
}

// ParseVersion parses a version string into a Version struct
func ParseVersion(version string) (*Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	// Initialize a Version struct
	v := &Version{}
	var err error

	// Parse the major version part
	v.Major, err = strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Parse the feature version part
	v.Feature, err = strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid feature version: %s", parts[1])
	}

	// Split the maintenance and hotfix parts
	maintenanceHotfix := strings.Split(parts[2], "-h")

	// Parse the maintenance version part
	v.Maintenance, err = strconv.Atoi(maintenanceHotfix[0])
	if err != nil {
		return nil, fmt.Errorf("invalid maintenance version: %s", maintenanceHotfix[0])
	}

	// If there's a hotfix part, parse it as well
	if len(maintenanceHotfix) > 1 {
		v.Hotfix, err = strconv.Atoi(maintenanceHotfix[1])
		if err != nil {
			return nil, fmt.Errorf("invalid hotfix version: %s", maintenanceHotfix[1])
		}
	}

	// Return the parsed Version struct
	return v, nil
}

// IsLessThan compares two Version structs
func (v *Version) IsLessThan(other *Version) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Feature != other.Feature {
		return v.Feature < other.Feature
	}
	if v.Maintenance != other.Maintenance {
		return v.Maintenance < other.Maintenance
	}
	return v.Hotfix < other.Hotfix
}

func IsAffectedVersion(device map[string]string, isGlobalProtect bool) (bool, string, error) {
	major, _ := strconv.Atoi(device["parsed_version_major"])
	feature, _ := strconv.Atoi(device["parsed_version_feature"])
	maintenance, _ := strconv.Atoi(device["parsed_version_maintenance"])
	hotfix, _ := strconv.Atoi(device["parsed_version_hotfix"])

	// Check if the version is 11.2 or later
	if major > 11 || (major == 11 && feature >= 2) {
		return false, "", nil // Versions 11.2 and later are not affected
	}

	v := &Version{
		Major:       major,
		Feature:     feature,
		Maintenance: maintenance,
		Hotfix:      hotfix,
	}

	featureRelease := fmt.Sprintf("%d.%d", v.Major, v.Feature)
	if isGlobalProtect && (featureRelease == "10.2" || featureRelease == "11.0" || featureRelease == "11.1") {
		featureRelease += "-gp"
	}

	minVersions, ok := config.MinimumPatchedVersions[featureRelease]
	if !ok {
		// If the feature release is not in MinimumPatchedVersions
		if v.Major < 8 || (v.Major == 8 && v.Feature < 1) {
			return true, "8.1.0", nil // Versions earlier than 8.1 are considered affected
		}
		return false, "", fmt.Errorf("unknown feature release: %s", featureRelease)
	}

	for _, minVersion := range minVersions {
		if v.Maintenance < minVersion.Maintenance || (v.Maintenance == minVersion.Maintenance && v.Hotfix < minVersion.Hotfix) {
			minUpdateRelease := fmt.Sprintf("%s.%d-h%d", featureRelease, minVersion.Maintenance, minVersion.Hotfix)
			return true, minUpdateRelease, nil
		}
	}

	return false, "", nil
}

func SplitDevices(deviceList []map[string]string) (affected []map[string]string, unaffected []map[string]string, err error) {
	for _, device := range deviceList {
		isAffected, minUpdateRelease, err := IsAffectedVersion(device, false) // Assuming no Global Protect for now
		if err != nil {
			return nil, nil, fmt.Errorf("error checking device %s: %v", device["hostname"], err)
		}

		deviceCopy := make(map[string]string)
		for k, v := range device {
			deviceCopy[k] = v
		}

		if isAffected {
			deviceCopy["minimumUpdateRelease"] = minUpdateRelease
			affected = append(affected, deviceCopy)
		} else {
			deviceCopy["result"] = "Not affected" // Default result for unaffected devices
			unaffected = append(unaffected, deviceCopy)
		}
	}

	return affected, unaffected, nil
}
