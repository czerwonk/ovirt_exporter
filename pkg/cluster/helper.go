// SPDX-License-Identifier: MIT

package cluster

import (
	"sync"

	"fmt"

	"github.com/czerwonk/ovirt_exporter/pkg/client"
	log "github.com/sirupsen/logrus"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

// Get retrieves cluster information
func Get(id string, cl client.Client) (*Cluster, error) {
	path := fmt.Sprintf("clusters/%s", id)

	c := Cluster{}
	err := cl.GetAndParse(path, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// Name retrieves cluster name
func Name(id string, cl client.Client) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[id]; found {
		return n
	}

	h, err := Get(id, cl)
	if err != nil {
		log.Error(err)
		return ""
	}

	nameCache[id] = h.Name
	return h.Name
}
