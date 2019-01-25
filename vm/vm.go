package vm

type VMs struct {
	VMs []VM `xml:"vm"`
}

type VM struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name"`
	Host struct {
		ID string `xml:"id,attr"`
	} `xml:"host,omitempty"`
	Cluster struct {
		ID string `xml:"id,attr"`
	} `xml:"cluster,omitempty"`
	Status string `xml:"status"`
	CPU    struct {
		Topology struct {
			Cores   int `xml:"cores"`
			Sockets int `xml:"sockets"`
			Threads int `xml:"threads"`
		} `xml:"topology"`
	} `xml:"cpu"`
	HasIllegalImages bool `json:"has_illegal_images"`
}
