package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect      SelectMode = iota // select randomly
	RoundRobbinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	//
	Refresh() error
	//
	Update(servers []string) error
	//
	Get(mode SelectMode) (string, error)
	//
	GetAll() ([]string, error)
}

type MultiServerDiscovery struct {
	r       *rand.Rand   // generate random number
	lock    sync.RWMutex // protext following
	servers []string
	index   int // record the selected position for robbin algorithm
}

func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

// Refresh doesn't make sense for MultiServerDiscovery, so ignore it
func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServerDiscovery) Update(servers []string) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.servers = servers
	return nil
}

// Get a server according to mode
func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobbinSelect:
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	}
	return "", errors.New("rpc discovery: not support select mode")
}

// return all servers in discovery
func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
