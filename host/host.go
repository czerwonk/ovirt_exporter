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
}
