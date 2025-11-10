package types

// SystemInfo contains operating system information
type SystemInfo struct {
	OSName    string `json:"os_name"`    // "Windows", "Darwin", "Linux"
	OSVersion string `json:"os_version"` // e.g., "10.0.19045", "14.1", "6.2.0"
	OSBuild   string `json:"os_build"`   // Build number or codename
	Timezone  string `json:"timezone"`   // IANA timezone (e.g., "America/New_York")
	Hostname  string `json:"hostname"`   // Machine hostname
}

// NetworkInfo contains network configuration
type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"` // Sorted by name
	WiFiSSIDs  []string           `json:"wifi_ssids"` // Known SSIDs, sorted
}

// NetworkInterface represents a single network adapter
type NetworkInterface struct {
	Name       string `json:"name"`        // Interface name
	IPAddress  string `json:"ip_address"`  // Primary IP (IPv4 or IPv6)
	MACAddress string `json:"mac_address"` // MAC address
}

// HardwareInfo contains hardware identifiers
type HardwareInfo struct {
	SerialNumber string `json:"serial_number"` // Machine serial number
	HardwareUUID string `json:"hardware_uuid"` // Hardware UUID
}

// PIIInfo contains personally identifiable information
type PIIInfo struct {
	Users          []User        `json:"users"`           // Local user accounts, sorted by username
	LoggedInUsers  []string      `json:"logged_in"`       // Currently logged in users, sorted
	HomeDirs       []string      `json:"home_dirs"`       // Home directory paths, sorted
	RecentProfiles []UserProfile `json:"recent_profiles"` // Recent login activity, sorted by timestamp
	PrimaryEmail   string        `json:"primary_email"`   // Best-effort email detection
}

// User represents a local user account
type User struct {
	Username string `json:"username"`
	FullName string `json:"full_name,omitempty"` // Display name
	UID      string `json:"uid,omitempty"`       // Unix UID or Windows SID
}

// UserProfile represents login activity
type UserProfile struct {
	Username   string `json:"username"`
	LastLogon  string `json:"last_logon"`            // ISO 8601 timestamp
	LogonCount int    `json:"logon_count,omitempty"` // Windows only
}
