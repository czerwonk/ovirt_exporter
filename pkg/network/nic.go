// SPDX-License-Identifier: MIT

package network

// HostNics is a collection of NICs of a host
type HostNics struct {
	Nics []Nic `xml:"host_nic"`
}

// VMNics is a collection of NICs of a VM
type VMNics struct {
	Nics []Nic `xml:"nic"`
}

// Nic represents the network interface card resource
type Nic struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name"`
	Mac  struct {
		Address string `xml:"address"`
	} `xml:"mac"`
}
