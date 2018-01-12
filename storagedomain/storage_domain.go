package storagedomain

type StorageDomains struct {
	Domains []StorageDomain `xml:"storage_domain"`
}

type StorageDomain struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name"`
	Storage struct {
		Path string `xml:"path"`
		Type string `xml:"type"`
	} `xml:"storage,omitempty"`
	Type           string  `xml:"type"`
	Available      float64 ` xml:"available"`
	Committed      float64 `xml:"committed"`
	Used           float64 `xml:"used"`
	ExternalStatus string  `xml:"external_status"`
	Master         bool    `xml:"master"`
	DataCenters    struct {
		DataCenter struct {
			ID string `xml:"id,attr"`
		} `xml:"data_center"`
	} `xml:"data_centers"`
}
