// SPDX-License-Identifier: MIT

package vm

// VMs is a collection of virtual machines
type VMs struct {
	VMs []VM `xml:"vm"`
}

// VM represents the virutal machine resource
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
	HasIllegalImages bool `xml:"has_illegal_images"`
}
