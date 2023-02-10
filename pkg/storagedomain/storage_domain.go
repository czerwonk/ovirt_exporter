// SPDX-License-Identifier: MIT

package storagedomain

// StorageDomains is a collection of storage domains
type StorageDomains struct {
	Domains []StorageDomain `xml:"storage_domain"`
}

// StorageDomain represents the storage domain resource
type StorageDomain struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,omitempty"`
	Storage struct {
		Path string `xml:"path,omitempty"`
		Type string `xml:"type,omitempty"`
	} `xml:"storage,omitempty"`
	Type           string  `xml:"type,omitempty"`
	Available      float64 ` xml:"available,omitempty"`
	Committed      float64 `xml:"committed,omitempty"`
	Used           float64 `xml:"used,omitempty"`
	ExternalStatus string  `xml:"external_status,omitempty"`
	Master         bool    `xml:"master,omitempty"`
	DataCenters    struct {
		DataCenter struct {
			ID string `xml:"id,attr"`
		} `xml:"data_center"`
	} `xml:"data_centers,omitempty"`
}
