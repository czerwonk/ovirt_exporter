// SPDX-License-Identifier: MIT

package disk

import (
	"context"
	"fmt"

	"github.com/czerwonk/ovirt_exporter/pkg/collector"
	"github.com/czerwonk/ovirt_exporter/pkg/storagedomain"
)

// Get retrieves disk information
func Get(ctx context.Context, id string, cl collector.Client) (*Disk, error) {
	path := fmt.Sprintf("disks/%s", id)

	d := &Disk{}
	err := cl.GetAndParse(ctx, path, &d)
	if err != nil {
		return nil, err
	}

	for i, dom := range d.StorageDomains.Domains {
		d.StorageDomains.Domains[i] = storagedomain.StorageDomain{
			ID:   dom.ID,
			Name: storagedomain.Name(ctx, dom.ID, cl),
		}
	}

	return d, nil
}
