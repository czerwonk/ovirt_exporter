// SPDX-License-Identifier: MIT

package host

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

// Get retrieves host information
func Get(id string, client *api.Client) (*Host, error) {
	path := fmt.Sprintf("hosts/%s", id)

	h := Host{}
	err := client.GetAndParse(path, &h)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

// Name retrieves host name
func Name(id string, client *api.Client) string {
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
