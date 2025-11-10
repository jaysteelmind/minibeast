package platform

import (
	"context"

	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Collector defines the platform-specific data collection interface
// Mathematical contract: All implementations must satisfy these operations
type Collector interface {
	// GetSystemInfo retrieves OS name, version, build, timezone
	// Complexity: O(1) - direct system calls
	// Timeout: Must respect context deadline
	GetSystemInfo(ctx context.Context) (*types.SystemInfo, error)

	// GetNetworkInfo retrieves IP addresses, MAC addresses, WiFi SSIDs
	// Complexity: O(n) where n = number of network interfaces
	// Timeout: Must respect context deadline
	GetNetworkInfo(ctx context.Context) (*types.NetworkInfo, error)

	// GetHardwareInfo retrieves serial numbers and hardware UUIDs
	// Complexity: O(1) - direct system queries
	// Timeout: Must respect context deadline
	GetHardwareInfo(ctx context.Context) (*types.HardwareInfo, error)

	// GetPIIInfo retrieves user accounts and logged-in users
	// Complexity: O(u) where u = number of users
	// Timeout: Must respect context deadline
	GetPIIInfo(ctx context.Context) (*types.PIIInfo, error)
}

// New creates a platform-specific collector for the current OS
// Mathematical guarantee: Returns non-nil collector or error
// Complexity: O(1)
func New() (Collector, error) {
	return newCollector()
}
