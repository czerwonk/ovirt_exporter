// SPDX-License-Identifier: MIT

package cluster

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

// Get retrieves cluster information
func Get(id string, client *api.Client) (*Cluster, error) {
	path := fmt.Sprintf("clusters/%s", id)

	c := Cluster{}
	err := client.GetAndParse(path, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Name retrieves cluster name
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
