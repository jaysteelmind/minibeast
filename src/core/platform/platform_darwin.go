//go:build darwin

package platform

import "github.com/minibeast/usb-agent/src/core/platform/darwin"

func newCollector() (Collector, error) {
	return darwin.NewCollector()
}
