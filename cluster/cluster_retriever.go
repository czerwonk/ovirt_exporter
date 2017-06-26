package cluster

import (
	"fmt"

	"sync"

	"github.com/czerwonk/ovirt_exporter/api"
)

var (
	cache map[string]*Cluster
	mutex = sync.Mutex{}
)

type ClusterRetriever struct {
	api *api.ApiClient
}

func init() {
	cache = make(map[string]*Cluster)
}

func NewRetriever(api *api.ApiClient) *ClusterRetriever {
	return &ClusterRetriever{api: api}
}

func (c *ClusterRetriever) Get(id string) (*Cluster, error) {
	mutex.Lock()
	defer mutex.Unlock()

	if c, found := cache[id]; found {
		return c, nil
	}

	var cluster Cluster
	err := c.api.GetAndParse(fmt.Sprintf("clusters/%s", id), &cluster)
	if err != nil {
		return nil, err
	}
	cache[id] = &cluster

	return &cluster, nil
}
