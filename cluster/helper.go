package cluster

import (
	"sync"

	"fmt"

	"github.com/czerwonk/ovirt_exporter/api"
	"github.com/prometheus/common/log"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

func Get(id string, client *api.ApiClient) (*Cluster, error) {
	path := fmt.Sprintf("clusters/%s", id)

	c := Cluster{}
	err := client.GetAndParse(path, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func Name(id string, client *api.ApiClient) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[id]; found {
		return n
	}

	h, err := Get(id, client)
	if err != nil {
		log.Error(err)
		return ""
	}

	nameCache[id] = h.Name
	return h.Name
}
