package config

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinimumPatchedVersions(t *testing.T) {
	// Test that all expected PAN-OS versions are present
	expectedVersions := []string{
		"8.1", "9.0", "9.1", "10.0", "10.1", "10.2", "10.2-gp",
		"11.0", "11.0-gp", "11.1", "11.1-gp",
	}

	for _, version := range expectedVersions {
		assert.Contains(t, MinimumPatchedVersions, version, "Expected version %s not found in MinimumPatchedVersions", version)
	}

	// Test that there are no unexpected versions
	assert.Len(t, MinimumPatchedVersions, len(expectedVersions), "Unexpected number of versions in MinimumPatchedVersions")

	// Test the structure and values for each version
	for version, patchVersions := range MinimumPatchedVersions {
		t.Run(version, func(t *testing.T) {
			assert.NotEmpty(t, patchVersions, "Patch versions for %s should not be empty", version)

			// Check that maintenance versions are in ascending order
			assert.True(t, sort.SliceIsSorted(patchVersions, func(i, j int) bool {
				return patchVersions[i].Maintenance < patchVersions[j].Maintenance
			}), "Maintenance versions for %s are not in ascending order", version)

			// Check that hotfix versions are non-negative
			for _, pv := range patchVersions {
				assert.GreaterOrEqual(t, pv.Maintenance, 0, "Maintenance version should be non-negative")
				assert.GreaterOrEqual(t, pv.Hotfix, 0, "Hotfix version should be non-negative")
			}

			// Additional checks for specific versions
			switch version {
			case "8.1":
				assert.Len(t, patchVersions, 3)
				assert.Equal(t, 21, patchVersions[0].Maintenance)
				assert.Equal(t, 3, patchVersions[0].Hotfix)
			case "11.1-gp":
				assert.Len(t, patchVersions, 3)
				assert.Equal(t, 0, patchVersions[0].Maintenance)
				assert.Equal(t, 3, patchVersions[0].Hotfix)
			}
		})
	}
}

func TestMinimumPatchedVersionStruct(t *testing.T) {
	mpv := MinimumPatchedVersion{Maintenance: 5, Hotfix: 3}

	assert.Equal(t, 5, mpv.Maintenance, "Unexpected Maintenance value")
	assert.Equal(t, 3, mpv.Hotfix, "Unexpected Hotfix value")
}
