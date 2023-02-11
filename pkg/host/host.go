// SPDX-License-Identifier: MIT

package host

// Hosts is a collection of Host
type Hosts struct {
	Hosts []Host `xml:"host"`
}

// Host represents the host resource
type Host struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name"`
	Cluster struct {
		ID string `xml:"id,attr"`
	} `xml:"cluster"`
	Status string `xml:"status"`
	CPU    struct {
		Speed    int `xml:"speed"`
		Topology struct {
			Cores   int `xml:"cores"`
			Sockets int `xml:"sockets"`
			Threads int `xml:"threads"`
		} `xml:"topology"`
	} `xml:"cpu"`
	Memory int64 `xml:"memory"`
}
