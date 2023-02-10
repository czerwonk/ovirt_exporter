// SPDX-License-Identifier: MIT

package disk

import (
	"fmt"

	"github.com/czerwonk/ovirt_exporter/pkg/client"
	"github.com/czerwonk/ovirt_exporter/pkg/storagedomain"
)

// Get retrieves disk information
func Get(id string, cl client.Client) (*Disk, error) {
	path := fmt.Sprintf("disks/%s", id)

	d := &Disk{}
	err := cl.GetAndParse(path, &d)
	if err != nil {
		return nil, err
	}

	for i, dom := range d.StorageDomains.Domains {
		d.StorageDomains.Domains[i] = storagedomain.StorageDomain{
			ID:   dom.ID,
			Name: storagedomain.Name(dom.ID, cl),
		}
	}

	return d, nil
}
