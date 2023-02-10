// SPDX-License-Identifier: MIT

package storagedomain

import (
	"context"
	"sync"

	"fmt"

	"github.com/czerwonk/ovirt_exporter/pkg/collector.go"
	log "github.com/sirupsen/logrus"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

// Get retrieves domain information
func Get(ctx context.Context, id string, cl collector.Client) (*StorageDomain, error) {
	path := fmt.Sprintf("storagedomains/%s", id)

	d := StorageDomain{}
	err := cl.GetAndParse(ctx, path, &d)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// Name retrieves domain name
func Name(ctx context.Context, id string, cl collector.Client) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[id]; found {
		return n
	}

	d, err := Get(ctx, id, cl)
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
