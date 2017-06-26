package cluster

type Cluster struct {
	Id          string `xml:"id,attr"`
	Name        string `xml:"name"`
	Description string `xml:"description"`
}
