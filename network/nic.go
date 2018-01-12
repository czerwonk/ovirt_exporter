package network

type HostNics struct {
	Nics []Nic `xml:"host_nic"`
}

type VMNics struct {
	Nics []Nic `xml:"nic"`
}

type Nic struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name"`
	Mac  struct {
		Address string `xml:"address"`
	} `xml:"mac"`
}
