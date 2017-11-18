package host

import (
	"sync"

	"fmt"

	"github.com/czerwonk/ovirt_api"
	"github.com/prometheus/common/log"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

func Get(id string, client *ovirt_api.ApiClient) (*Host, error) {
	path := fmt.Sprintf("hosts/%s", id)

	h := Host{}
	err := client.GetAndParse(path, &h)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

func Name(id string, client *ovirt_api.ApiClient) string {
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
