package storagedomain

import (
	"sync"

	"fmt"

	"github.com/czerwonk/ovirt_api/api"
	log "github.com/sirupsen/logrus"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

// Get retrieves domain information
func Get(id string, client *api.Client) (*StorageDomain, error) {
	path := fmt.Sprintf("storagedomains/%s", id)

	d := StorageDomain{}
	err := client.GetAndParse(path, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// Name retrieves domain name
func Name(id string, client *api.Client) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[id]; found {
		return n
	}

	d, err := Get(id, client)
	if err != nil {
		log.Error(err)
		return ""
	}

	if d == nil {
		log.Errorf("could not find name for storage domain with ID %s", id)
		return ""
	}

	nameCache[id] = d.Name
	return d.Name
}
