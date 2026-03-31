package discoversrv

import "fmt"

// DiscoveredService represents a WebDAV service found on the local network
// via mDNS/DNS-SD.
type DiscoveredService struct {
	// identity
	Name   string
	Host   string

	// connection
	Address string
	Port    uint16
}

// URL returns the WebDAV URL for this service.
func (receiver DiscoveredService) URL() string {
	return fmt.Sprintf("http://%s:%d/", receiver.Address, receiver.Port)
}
