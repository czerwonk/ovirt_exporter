package host

import (
	"fmt"

	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
)

var (
	cache map[string]*Host
	mutex = sync.Mutex{}
)

type HostRetriever struct {
	api *api.ApiClient
}

func init() {
	cache = make(map[string]*Host)
}

func NewRetriever(api *api.ApiClient) *HostRetriever {
	return &HostRetriever{api: api}
}

func (c *HostRetriever) Get(id string) (*Host, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if h, found := cache[id]; found {
		return h, nil
	}

	var host Host
	err := c.api.GetAndParse(fmt.Sprintf("hosts/%s", id), &host)
	if err != nil {
		return nil, err
	}
	cache[id] = &host

	return &host, nil
}
