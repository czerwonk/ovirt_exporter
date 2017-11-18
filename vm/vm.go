package vm

type Vms struct {
	Vm []Vm `xml:"vm"`
}

type Vm struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name"`
	Host struct {
		Id string `xml:"id,attr"`
	} `xml:"host,omitempty"`
	Cluster struct {
		Id string `xml:"id,attr"`
	} `xml:"cluster,omitempty"`
	Status string `xml:"status"`
	Cpu    struct {
		Topology struct {
			Cores   int `xml:"cores"`
			Sockets int `xml:"sockets"`
			Threads int `xml:"threads"`
		} `xml:"topology"`
	} `xml:"cpu"`
}
