package datacenter

import (
	"fmt"

	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
)

var (
	cache map[string]*DataCenter
	mutex = sync.Mutex{}
)

type DatacenterRetriever struct {
	api *api.ApiClient
}

func init() {
	cache = make(map[string]*DataCenter)
}

func NewRetriever(api *api.ApiClient) *DatacenterRetriever {
	return &DatacenterRetriever{api: api}
}

func (c *DatacenterRetriever) Get(id string) (*DataCenter, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if dc, found := cache[id]; found {
		return dc, nil
	}

	var dc DataCenter
	err := c.api.GetAndParse(fmt.Sprintf("datacenters/%s", id), &dc)
	if err != nil {
		return nil, err
	}
	cache[id] = &dc

	return &dc, nil
}
