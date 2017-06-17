package host

type Hosts struct {
	Host []Host `xml:"host"`
}

type Host struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name"`
}
