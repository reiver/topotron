package discoversrv

import (
	"sync"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/godbus/dbus/v5"
	avahi "github.com/holoplot/go-avahi"

	"topotron/srv/log"
)

const serviceType = "_webdav._tcp"

// Discovery browses for WebDAV services on the local network via Avahi D-Bus.
type Discovery struct {
	mu       sync.Mutex
	services map[string]DiscoveredService
	stopCh   chan struct{}

	// callbacks (called on the GTK main thread via glib.IdleAdd)
	OnAdded   func(service DiscoveredService)
	OnRemoved func(name string)
}

// New creates a new [Discovery] and starts browsing in a background goroutine.
func New() *Discovery {
	receiver := &Discovery{
		services: make(map[string]DiscoveredService),
		stopCh:   make(chan struct{}),
	}

	go receiver.browse()

	return receiver
}

// Services returns a snapshot of currently discovered services.
func (receiver *Discovery) Services() []DiscoveredService {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	result := make([]DiscoveredService, 0, len(receiver.services))
	for _, svc := range receiver.services {
		result = append(result, svc)
	}

	return result
}

// Stop stops the background browsing.
func (receiver *Discovery) Stop() {
	close(receiver.stopCh)
}

func (receiver *Discovery) browse() {
	log := logsrv.Begin()
	defer log.End()

	conn, err := dbus.SystemBus()
	if nil != err {
		log.Highlightf("could not connect to D-Bus: %v", err)
		return
	}

	server, err := avahi.ServerNew(conn)
	if nil != err {
		log.Highlightf("could not connect to Avahi: %v", err)
		return
	}

	sb, err := server.ServiceBrowserNew(
		avahi.InterfaceUnspec,
		avahi.ProtoUnspec,
		serviceType,
		"local",
		0,
	)
	if nil != err {
		log.Highlightf("could not create Avahi service browser: %v", err)
		return
	}

	log.Highlightf("browsing for %s services", serviceType)

	for {
		select {
		case <-receiver.stopCh:
			return

		case svc := <-sb.AddChannel:
			receiver.onServiceAdded(server, svc)

		case svc := <-sb.RemoveChannel:
			receiver.onServiceRemoved(svc)
		}
	}
}

func (receiver *Discovery) onServiceAdded(server *avahi.Server, svc avahi.Service) {
	log := logsrv.Begin()
	defer log.End()

	resolved, err := server.ResolveService(
		svc.Interface,
		svc.Protocol,
		svc.Name,
		svc.Type,
		svc.Domain,
		avahi.ProtoUnspec,
		0,
	)
	if nil != err {
		log.Highlightf("could not resolve service %s: %v", svc.Name, err)
		return
	}

	discovered := DiscoveredService{
		Name:    resolved.Name,
		Host:    resolved.Host,
		Address: resolved.Address,
		Port:    resolved.Port,
	}

	receiver.mu.Lock()
	receiver.services[discovered.Name] = discovered
	receiver.mu.Unlock()

	log.Highlightf("discovered: %s at %s:%d", discovered.Name, discovered.Address, discovered.Port)

	if nil != receiver.OnAdded {
		svcCopy := discovered
		glib.IdleAdd(func() {
			receiver.OnAdded(svcCopy)
		})
	}
}

func (receiver *Discovery) onServiceRemoved(svc avahi.Service) {
	log := logsrv.Begin()
	defer log.End()

	receiver.mu.Lock()
	delete(receiver.services, svc.Name)
	receiver.mu.Unlock()

	log.Highlightf("removed: %s", svc.Name)

	if nil != receiver.OnRemoved {
		name := svc.Name
		glib.IdleAdd(func() {
			receiver.OnRemoved(name)
		})
	}
}
