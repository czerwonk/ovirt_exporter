package datacenter

type DataCenter struct {
	Id          string `xml:"id,attr"`
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Status      string `xml:"status"`
}
