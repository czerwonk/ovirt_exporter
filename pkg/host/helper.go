// SPDX-License-Identifier: MIT

package host

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

// Get retrieves host information
func Get(ctx context.Context, id string, cl collector.Client) (*Host, error) {
	path := fmt.Sprintf("hosts/%s", id)

	h := Host{}
	err := cl.GetAndParse(ctx, path, &h)
	if err != nil {
		return nil, err
	}

	return &h, nil
}

// Name retrieves host name
func Name(ctx context.Context, id string, cl collector.Client) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[id]; found {
		return n
	}

	h, err := Get(ctx, id, cl)
	if err != nil {
		log.Error(err)
		return ""
	}

	nameCache[id] = h.Name
	return h.Name
}
