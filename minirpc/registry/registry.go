package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type MiniRegistry struct {
	timeout time.Duration
	lock    sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath   = "/_minirpc_/registry"
	defaultHeader = "X-Minirpc-Servers"
	defaultTime   = time.Minute * 5
)

func New(timeout time.Duration) *MiniRegistry {
	return &MiniRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultMiniRegistry = New(defaultTime)

// add server to registry
func (r *MiniRegistry) postServer(addr string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		s.start = time.Now()
	}
}

// get usable servers
func (r *MiniRegistry) GetServers() []string {
	r.lock.Lock()
	defer r.lock.Unlock()

	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	log.Println("servers ", alive)
	sort.Strings(alive)
	return alive
}

func (r *MiniRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set(defaultHeader, strings.Join(r.GetServers(), ","))
	case "POST":
		addrs := req.Header.Get(defaultHeader)
		if addrs == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.postServer(addrs)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *MiniRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultMiniRegistry.HandleHTTP(defaultPath)
}

// var _ MiniRegistry = (*http.Server)(nil)

// HeartBeat send a heardbeat message every once in a while
// it's a helper function for a server to register or send heartbeat
func HeartBeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		// make sure there is enough time to send heart beat
		// before it's removed from registry
		duration = defaultTime - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartBeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartBeat(registry, addr)
		}
	}()
}

func sendHeartBeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry ", registry)
	httpClient := &http.Client{}
	r, _ := http.NewRequest("POST", registry, nil)
	r.Header.Set(defaultHeader, addr)
	if _, err := httpClient.Do(r); err != nil {
		log.Println("rpc servber: heartbeat error: ", err)
		return err
	}
	return nil
}
