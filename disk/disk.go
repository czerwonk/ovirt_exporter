package disk

import "github.com/czerwonk/ovirt_exporter/storagedomain"

// Disk represents the disk resource
type Disk struct {
	ID              string                        `xml:"id,attr"`
	Name            string                        `xml:"name,omitempty"`
	Alias           string                        `xml:"alias,omitempty"`
	ProvisionedSize uint64                        `xml:"provisioned_size,omitempty"`
	ActualSize      uint64                        `xml:"actual_size,omitempty"`
	TotalSize       uint64                        `xml:"total_size,omitempty"`
	StorageDomains  *storagedomain.StorageDomains `xml:"storage_domains,omitempty"`
}

// StorageDomainName returns the name of the storage domain of the disk
func (d *Disk) StorageDomainName() string {
	if len(d.StorageDomains.Domains) == 0 {
		return ""
	}

	return d.StorageDomains.Domains[0].Name
}
