package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAffectedFamilies(t *testing.T) {
	assert.Contains(t, AffectedFamilies, "200")
	assert.Contains(t, AffectedFamilies["200"], "PA-200")

	assert.Contains(t, AffectedFamilies, "3200")
	assert.Contains(t, AffectedFamilies["3200"], "PA-3260")

	assert.Contains(t, AffectedFamilies, "vm")
	assert.Contains(t, AffectedFamilies["vm"], "PA-VM")
	assert.Contains(t, AffectedFamilies["vm"], "PA-VM (lite)")

	assert.NotContains(t, AffectedFamilies, "400")
}

func TestUnaffectedFamilies(t *testing.T) {
	assert.Contains(t, UnaffectedFamilies, "400")
	assert.Contains(t, UnaffectedFamilies["400"], "PA-410")
	assert.Contains(t, UnaffectedFamilies["400"], "PA-460")

	assert.Contains(t, UnaffectedFamilies, "5400f")
	assert.Contains(t, UnaffectedFamilies["5400f"], "PA-5440")

	assert.Contains(t, UnaffectedFamilies, "7500")
	assert.Contains(t, UnaffectedFamilies["7500"], "PA-7500")

	assert.NotContains(t, UnaffectedFamilies, "200")
}

func TestFamilyCompleteness(t *testing.T) {
	allFamilies := make(map[string]bool)

	for family := range AffectedFamilies {
		allFamilies[family] = true
	}

	for family := range UnaffectedFamilies {
		allFamilies[family] = true
	}

	expectedFamilies := []string{
		"200", "220", "3000", "3200", "500", "5000", "5200", "7000", "7000b", "800", "vm", "vmarm",
		"400", "1400", "3400", "5400", "5400f", "7500",
	}

	for _, family := range expectedFamilies {
		assert.True(t, allFamilies[family], "Family %s should be in either AffectedFamilies or UnaffectedFamilies", family)
	}

	assert.Equal(t, len(expectedFamilies), len(allFamilies), "There should be no unexpected families")
}
