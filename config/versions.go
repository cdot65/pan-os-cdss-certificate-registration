package config

// MinimumPatchedVersion represents the minimum patched version for a specific release
type MinimumPatchedVersion struct {
	Maintenance int
	Hotfix      int
}

// MinimumPatchedVersions represents the minimum patched versions for each PAN-OS feature release
var MinimumPatchedVersions = map[string][]MinimumPatchedVersion{
	"8.1": {
		{Maintenance: 21, Hotfix: 3},
		{Maintenance: 25, Hotfix: 3},
		{Maintenance: 26, Hotfix: 0},
	},
	"9.0": {
		{Maintenance: 16, Hotfix: 7},
		{Maintenance: 17, Hotfix: 5},
	},
	"9.1": {
		{Maintenance: 11, Hotfix: 5},
		{Maintenance: 12, Hotfix: 7},
		{Maintenance: 13, Hotfix: 5},
		{Maintenance: 14, Hotfix: 8},
		{Maintenance: 16, Hotfix: 5},
		{Maintenance: 17, Hotfix: 0},
	},
	"10.0": {
		{Maintenance: 8, Hotfix: 8},
		{Maintenance: 11, Hotfix: 4},
		{Maintenance: 12, Hotfix: 5},
	},
	"10.1": {
		{Maintenance: 3, Hotfix: 3},
		{Maintenance: 4, Hotfix: 6},
		{Maintenance: 5, Hotfix: 4},
		{Maintenance: 6, Hotfix: 8},
		{Maintenance: 7, Hotfix: 1},
		{Maintenance: 8, Hotfix: 7},
		{Maintenance: 9, Hotfix: 8},
		{Maintenance: 10, Hotfix: 5},
		{Maintenance: 11, Hotfix: 5},
		{Maintenance: 12, Hotfix: 0},
	},
	"10.2": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 1},
		{Maintenance: 2, Hotfix: 4},
		{Maintenance: 3, Hotfix: 12},
		{Maintenance: 4, Hotfix: 10},
		{Maintenance: 6, Hotfix: 1},
		{Maintenance: 7, Hotfix: 3},
		{Maintenance: 8, Hotfix: 0},
	},
	"10.2-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 2},
		{Maintenance: 2, Hotfix: 5},
		{Maintenance: 3, Hotfix: 13},
		{Maintenance: 4, Hotfix: 16},
		{Maintenance: 5, Hotfix: 6},
		{Maintenance: 6, Hotfix: 3},
		{Maintenance: 7, Hotfix: 8},
		{Maintenance: 8, Hotfix: 3},
		{Maintenance: 9, Hotfix: 1},
	},
	"11.0": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 3},
		{Maintenance: 2, Hotfix: 3},
		{Maintenance: 3, Hotfix: 3},
		{Maintenance: 4, Hotfix: 0},
	},
	"11.0-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 4},
		{Maintenance: 2, Hotfix: 4},
		{Maintenance: 3, Hotfix: 10},
		{Maintenance: 4, Hotfix: 1},
	},
	"11.1": {
		{Maintenance: 0, Hotfix: 2},
		{Maintenance: 1, Hotfix: 0},
	},
	"11.1-gp": {
		{Maintenance: 0, Hotfix: 3},
		{Maintenance: 1, Hotfix: 1},
		{Maintenance: 2, Hotfix: 3},
	},
}
