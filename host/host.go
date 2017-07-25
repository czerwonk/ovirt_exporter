package host

type Hosts struct {
	Host []Host `xml:"host"`
}

type Host struct {
	Id      string `xml:"id,attr"`
	Name    string `xml:"name"`
	Cluster struct {
		Id string `xml:"id,attr"`
	} `xml:"cluster"`
	Status string `xml:"status"`
	Cpu    struct {
		Speed    int `xml:"speed"`
		Topology struct {
			Cores   int `xml:"cores"`
			Sockets int `xml:"sockets"`
			Threads int `xml:"threads"`
		} `xml:"topology"`
	} `xml:"cpu"`
	Memory int64 `xml:"memory"`
}
