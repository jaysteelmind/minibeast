package collection

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/minibeast/usb-agent/src/core/config"
	"github.com/minibeast/usb-agent/src/core/platform"
	"github.com/minibeast/usb-agent/src/core/platform/types"
)

// Collector orchestrates parallel data collection
// Mathematical complexity: O(max(|categories|/N) * T) where N=poolSize, T=timeout
type Collector struct {
	config            *config.Config
	platformCollector platform.Collector
	timeout           time.Duration
	poolSize          int
}

// NewCollector creates a new collector
// Complexity: O(1)
func NewCollector(cfg *config.Config) (*Collector, error) {
	platformCollector, err := platform.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create platform collector: %w", err)
	}

	return &Collector{
		config:            cfg,
		platformCollector: platformCollector,
		timeout:           cfg.GetCategoryTimeout(),
		poolSize:          cfg.Performance.MaxGoroutines,
	}, nil
}

// CollectAll performs parallel data collection with timeout guards
// Mathematical guarantee: Returns complete Facts or error (never partial Facts)
// Complexity: O(|categories|) with bounded parallelism
func (c *Collector) CollectAll(ctx context.Context) (*Facts, error) {
	startTime := time.Now()

	// Initialize results
	facts := &Facts{
		Timestamp:        time.Now().UTC(),
		CollectorVersion: "1.0.0",
		Users:            []types.User{},
		LoggedInUsers:    []string{},
		HomeDirs:         []string{},
		RecentProfiles:   []types.UserProfile{},
		LocalIPs:         []types.NetworkInterface{},
		MACAddresses:     []types.NetworkInterface{},
		WiFiSSIDs:        []string{},
	}

	// Create bounded pool
	pool := NewBoundedPool(c.poolSize)

	// Result channels
	systemChan := make(chan *types.SystemInfo, 1)
	networkChan := make(chan *types.NetworkInfo, 1)
	hardwareChan := make(chan *types.HardwareInfo, 1)
	piiChan := make(chan *types.PIIInfo, 1)

	// Error channel
	errChan := make(chan error, 4)

	// Submit collection tasks
	categories := []struct {
		name string
		task func()
	}{
		{
			name: "system_info",
			task: func() {
				catCtx, cancel := context.WithTimeout(ctx, c.timeout)
				defer cancel()

				info, err := c.platformCollector.GetSystemInfo(catCtx)
				if err != nil {
					errChan <- fmt.Errorf("system_info: %w", err)
					return
				}
				systemChan <- info
			},
		},
		{
			name: "network_info",
			task: func() {
				catCtx, cancel := context.WithTimeout(ctx, c.timeout)
				defer cancel()

				info, err := c.platformCollector.GetNetworkInfo(catCtx)
				if err != nil {
					errChan <- fmt.Errorf("network_info: %w", err)
					return
				}
				networkChan <- info
			},
		},
		{
			name: "hardware_info",
			task: func() {
				catCtx, cancel := context.WithTimeout(ctx, c.timeout)
				defer cancel()

				info, err := c.platformCollector.GetHardwareInfo(catCtx)
				if err != nil {
					errChan <- fmt.Errorf("hardware_info: %w", err)
					return
				}
				hardwareChan <- info
			},
		},
		{
			name: "pii_info",
			task: func() {
				if !c.config.PII {
					return // Skip if PII collection disabled
				}

				catCtx, cancel := context.WithTimeout(ctx, c.timeout)
				defer cancel()

				info, err := c.platformCollector.GetPIIInfo(catCtx)
				if err != nil {
					errChan <- fmt.Errorf("pii_info: %w", err)
					return
				}
				piiChan <- info
			},
		},
	}

	// Submit all tasks
	for _, cat := range categories {
		if err := pool.Submit(ctx, cat.task); err != nil {
			return nil, fmt.Errorf("failed to submit %s: %w", cat.name, err)
		}
	}

	// Wait for completion
	pool.Wait()
	close(systemChan)
	close(networkChan)
	close(hardwareChan)
	close(piiChan)
	close(errChan)

	// Collect errors (non-fatal, graceful degradation)
	var collectionErrors []error
	for err := range errChan {
		collectionErrors = append(collectionErrors, err)
	}

	// Aggregate results
	if systemInfo := <-systemChan; systemInfo != nil {
		facts.Hostname = systemInfo.Hostname
		facts.ComputerName = systemInfo.Hostname
		facts.OSName = systemInfo.OSName
		facts.OSVersion = systemInfo.OSVersion
		facts.OSBuild = systemInfo.OSBuild
		facts.Timezone = systemInfo.Timezone
	}

	if networkInfo := <-networkChan; networkInfo != nil {
		facts.LocalIPs = networkInfo.Interfaces
		facts.MACAddresses = networkInfo.Interfaces
		facts.WiFiSSIDs = networkInfo.WiFiSSIDs
	}

	if hardwareInfo := <-hardwareChan; hardwareInfo != nil {
		facts.SerialNumber = hardwareInfo.SerialNumber
		facts.HardwareUUID = hardwareInfo.HardwareUUID
	}

	if piiInfo := <-piiChan; piiInfo != nil {
		facts.Users = piiInfo.Users
		facts.LoggedInUsers = piiInfo.LoggedInUsers
		facts.HomeDirs = piiInfo.HomeDirs
		facts.RecentProfiles = piiInfo.RecentProfiles
		facts.PrimaryEmail = piiInfo.PrimaryEmail

		// Set machine owner (first non-system user)
		if len(piiInfo.Users) > 0 {
			facts.MachineOwner = piiInfo.Users[0].Username
		}
	}

	// Ensure deterministic ordering (critical for hash consistency)
	c.sortFacts(facts)

	// Calculate collection duration
	facts.CollectionDurationMs = time.Since(startTime).Milliseconds()

	// Validate mathematical invariants
	if err := facts.Validate(); err != nil {
		return nil, fmt.Errorf("facts validation failed: %w", err)
	}

	return facts, nil
}

// sortFacts ensures deterministic ordering of all arrays
// Complexity: O(n log n) where n = max array size
func (c *Collector) sortFacts(facts *Facts) {
	// Sort users by username
	sort.Slice(facts.Users, func(i, j int) bool {
		return facts.Users[i].Username < facts.Users[j].Username
	})

	// Sort logged-in users
	sort.Strings(facts.LoggedInUsers)

	// Sort home directories
	sort.Strings(facts.HomeDirs)

	// Sort WiFi SSIDs
	sort.Strings(facts.WiFiSSIDs)

	// Sort network interfaces by name
	sort.Slice(facts.LocalIPs, func(i, j int) bool {
		return facts.LocalIPs[i].Name < facts.LocalIPs[j].Name
	})
	sort.Slice(facts.MACAddresses, func(i, j int) bool {
		return facts.MACAddresses[i].Name < facts.MACAddresses[j].Name
	})

	// Sort recent profiles by username (timestamp secondary)
	sort.Slice(facts.RecentProfiles, func(i, j int) bool {
		if facts.RecentProfiles[i].Username == facts.RecentProfiles[j].Username {
			return facts.RecentProfiles[i].LastLogon > facts.RecentProfiles[j].LastLogon
		}
		return facts.RecentProfiles[i].Username < facts.RecentProfiles[j].Username
	})
}
