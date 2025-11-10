package collection

import (
	"time"

	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Facts represents the complete system snapshot
// Mathematical invariant: All fields deterministic for given hardware state
type Facts struct {
	// Metadata
	Timestamp            time.Time `json:"timestamp"`              // ISO 8601 (UTC)
	CollectionDurationMs int64     `json:"collection_duration_ms"` // Performance tracking
	CollectorVersion     string    `json:"collector_version"`      // Version tracking

	// System identification
	Hostname     string `json:"hostname"`
	MachineOwner string `json:"machine_owner,omitempty"` // Best-effort
	ComputerName string `json:"computer_name"`

	// User information (sorted for determinism)
	Users          []types.User        `json:"users"`           // Sorted by username
	LoggedInUsers  []string            `json:"logged_in_users"` // Sorted
	HomeDirs       []string            `json:"home_dirs"`       // Sorted by path
	RecentProfiles []types.UserProfile `json:"recent_profiles"` // Sorted by timestamp
	PrimaryEmail   string              `json:"primary_user_email,omitempty"`

	// Network information (sorted for determinism)
	LocalIPs     []types.NetworkInterface `json:"local_ips"`        // Sorted by interface name
	MACAddresses []types.NetworkInterface `json:"mac_addresses"`    // Sorted by interface name
	WiFiSSIDs    []string                 `json:"wifi_known_ssids"` // Sorted

	// Hardware identifiers
	SerialNumber string `json:"serial_number"`
	HardwareUUID string `json:"hardware_uuid"`

	// Operating system
	OSName    string `json:"os_name"` // "Windows", "Darwin", "Linux"
	OSVersion string `json:"os_version"`
	OSBuild   string `json:"os_build"`
	Timezone  string `json:"timezone"` // IANA format
}

// Validate checks mathematical invariants
// Returns error if invariants violated
// Complexity: O(1)
func (f *Facts) Validate() error {
	// All critical fields must be non-empty
	if f.Hostname == "" {
		return &ValidationError{Field: "hostname", Reason: "must not be empty"}
	}
	if f.OSName == "" {
		return &ValidationError{Field: "os_name", Reason: "must not be empty"}
	}
	if f.HardwareUUID == "" {
		return &ValidationError{Field: "hardware_uuid", Reason: "must not be empty"}
	}

	return nil
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return "validation failed: " + e.Field + " - " + e.Reason
}

// Category represents a data collection category
type Category string

const (
	CategorySystemInfo   Category = "system_info"
	CategoryNetworkInfo  Category = "network_info"
	CategoryHardwareInfo Category = "hardware_info"
	CategoryPIIInfo      Category = "pii_info"
)
