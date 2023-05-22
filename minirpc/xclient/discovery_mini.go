package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type MiniRegistryDiscovery struct {
	*MultiServerDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewMiniRegistryDiscovery(registryAddr string, timeout time.Duration) *MiniRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	return &MiniRegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:             registryAddr,
		timeout:              timeout,
	}
}

func (d *MiniRegistryDiscovery) Update(servers []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *MiniRegistryDiscovery) Refresh() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}

	log.Println("rpc registry: refresh servers registry ", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry refresh error: ", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-Minirpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *MiniRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServerDiscovery.Get(mode)
}

func (d *MiniRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServerDiscovery.GetAll()
}
