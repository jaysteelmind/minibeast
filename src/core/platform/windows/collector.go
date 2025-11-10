package windows

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Collector implements platform.Collector for Windows systems
type Collector struct{}

// NewCollector creates a new Windows collector
// Complexity: O(1)
func NewCollector() (*Collector, error) {
	return &Collector{}, nil
}

// GetSystemInfo retrieves Windows system information
// Complexity: O(1)
func (c *Collector) GetSystemInfo(ctx context.Context) (*types.SystemInfo, error) {
	info := &types.SystemInfo{
		OSName: "Windows",
	}

	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	} else {
		info.Hostname = "unknown"
	}

	// Get Windows version using systeminfo
	if version, err := c.getWindowsVersion(); err == nil {
		info.OSVersion = version
	} else {
		info.OSVersion = "unknown"
	}

	// Get build number
	if build, err := c.getBuildNumber(); err == nil {
		info.OSBuild = build
	} else {
		info.OSBuild = "unknown"
	}

	// Get timezone
	info.Timezone = time.Local.String()

	return info, nil
}

// GetNetworkInfo retrieves Windows network configuration
// Complexity: O(n) where n = number of network interfaces
func (c *Collector) GetNetworkInfo(ctx context.Context) (*types.NetworkInfo, error) {
	info := &types.NetworkInfo{
		Interfaces: []types.NetworkInterface{},
		WiFiSSIDs:  []string{},
	}

	// Get network interfaces using ipconfig
	interfaces, err := c.getNetworkInterfaces()
	if err == nil {
		info.Interfaces = interfaces
	}

	// Get WiFi SSIDs using netsh
	ssids, err := c.getWiFiSSIDs()
	if err == nil {
		info.WiFiSSIDs = ssids
	}

	// Sort for determinism
	sort.Slice(info.Interfaces, func(i, j int) bool {
		return info.Interfaces[i].Name < info.Interfaces[j].Name
	})
	sort.Strings(info.WiFiSSIDs)

	return info, nil
}

// GetHardwareInfo retrieves Windows hardware identifiers
// Complexity: O(1)
func (c *Collector) GetHardwareInfo(ctx context.Context) (*types.HardwareInfo, error) {
	info := &types.HardwareInfo{
		SerialNumber: "unknown",
		HardwareUUID: "unknown",
	}

	// Get hardware UUID using wmic
	if uuid, err := c.getHardwareUUID(); err == nil {
		info.HardwareUUID = uuid
	}

	// Get serial number using wmic
	if serial, err := c.getSerialNumber(); err == nil {
		info.SerialNumber = serial
	}

	return info, nil
}

// GetPIIInfo retrieves Windows user information
// Complexity: O(u) where u = number of users
func (c *Collector) GetPIIInfo(ctx context.Context) (*types.PIIInfo, error) {
	info := &types.PIIInfo{
		Users:          []types.User{},
		LoggedInUsers:  []string{},
		HomeDirs:       []string{},
		RecentProfiles: []types.UserProfile{},
		PrimaryEmail:   "unknown",
	}

	// Get local users using wmic
	users, err := c.getLocalUsers()
	if err == nil {
		info.Users = users
		for _, u := range users {
			// Windows user home directories
			homeDir := "C:\\Users\\" + u.Username
			info.HomeDirs = append(info.HomeDirs, homeDir)
		}
	}

	// Get currently logged-in user
	currentUser, err := user.Current()
	if err == nil {
		info.LoggedInUsers = []string{currentUser.Username}
	}

	// Sort for determinism
	sort.Slice(info.Users, func(i, j int) bool {
		return info.Users[i].Username < info.Users[j].Username
	})
	sort.Strings(info.LoggedInUsers)
	sort.Strings(info.HomeDirs)

	return info, nil
}

// Helper functions

func (c *Collector) getWindowsVersion() (string, error) {
	cmd := exec.Command("cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output like "Microsoft Windows [Version 10.0.19045.1234]"
	version := strings.TrimSpace(string(output))
	if strings.Contains(version, "[Version ") {
		start := strings.Index(version, "[Version ") + 9
		end := strings.Index(version[start:], "]")
		if end > 0 {
			return version[start : start+end], nil
		}
	}

	return version, nil
}

func (c *Collector) getBuildNumber() (string, error) {
	// Use wmic to get build number
	cmd := exec.Command("wmic", "os", "get", "BuildNumber", "/value")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "BuildNumber=") {
			return strings.TrimPrefix(line, "BuildNumber="), nil
		}
	}

	return "", nil
}

func (c *Collector) getNetworkInterfaces() ([]types.NetworkInterface, error) {
	interfaces := []types.NetworkInterface{}

	// Use ipconfig /all to get network information
	cmd := exec.Command("ipconfig", "/all")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var currentIface *types.NetworkInterface

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// New adapter section
		if strings.HasSuffix(line, ":") && strings.Contains(line, "adapter") {
			if currentIface != nil && currentIface.Name != "" {
				interfaces = append(interfaces, *currentIface)
			}
			// Extract adapter name
			name := strings.TrimSuffix(line, ":")
			currentIface = &types.NetworkInterface{
				Name:       name,
				IPAddress:  "unknown",
				MACAddress: "unknown",
			}
		} else if currentIface != nil {
			// Parse IPv4 Address
			if strings.Contains(line, "IPv4 Address") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					ip := strings.TrimSpace(parts[1])
					// Remove (Preferred) suffix if present
					ip = strings.Split(ip, "(")[0]
					currentIface.IPAddress = strings.TrimSpace(ip)
				}
			}
			// Parse Physical Address (MAC)
			if strings.Contains(line, "Physical Address") {
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					mac := strings.TrimSpace(strings.Join(parts[1:], ":"))
					currentIface.MACAddress = mac
				}
			}
		}
	}

	if currentIface != nil && currentIface.Name != "" {
		interfaces = append(interfaces, *currentIface)
	}

	return interfaces, nil
}

func (c *Collector) getWiFiSSIDs() ([]string, error) {
	ssids := []string{}

	// Use netsh to get WiFi profiles
	cmd := exec.Command("netsh", "wlan", "show", "profiles")
	output, err := cmd.Output()
	if err != nil {
		return ssids, nil // Best-effort, not fatal
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "All User Profile") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ssid := strings.TrimSpace(parts[1])
				if ssid != "" {
					ssids = append(ssids, ssid)
				}
			}
		}
	}

	return ssids, nil
}

func (c *Collector) getHardwareUUID() (string, error) {
	cmd := exec.Command("wmic", "csproduct", "get", "UUID", "/value")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "UUID=") {
			return strings.TrimPrefix(line, "UUID="), nil
		}
	}

	return "", nil
}

func (c *Collector) getSerialNumber() (string, error) {
	cmd := exec.Command("wmic", "bios", "get", "serialnumber", "/value")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SerialNumber=") {
			return strings.TrimPrefix(line, "SerialNumber="), nil
		}
	}

	return "", nil
}

func (c *Collector) getLocalUsers() ([]types.User, error) {
	users := []types.User{}

	cmd := exec.Command("wmic", "useraccount", "get", "name,fullname,sid", "/format:csv")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i < 2 {
			continue // Skip header rows
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) >= 4 {
			username := strings.TrimSpace(fields[2])
			fullname := strings.TrimSpace(fields[1])
			sid := strings.TrimSpace(fields[3])

			// Filter out system accounts
			if username != "" && !strings.HasPrefix(username, "SYSTEM") {
				users = append(users, types.User{
					Username: username,
					FullName: fullname,
					UID:      sid,
				})
			}
		}
	}

	return users, nil
}
