package main

import (
	"flag"
	"log"

	"github.com/czerwonk/ovirt_api/api"
)

func main() {
	user := flag.String("user", "username", "username")
	pass := flag.String("pass", "password", "password")
	url := flag.String("url", "http://ovirt.engine", "api-url")
	flag.Parse()

	c, err := api.NewClient(*url, *user, *pass)
	if err != nil {
		log.Fatal(err)
	}

	vms := &VMs{}
	err = c.SendAndParse("/vms", "GET", vms, nil)
	if err != nil {
		log.Fatal(err)
	}

	for _, vm := range vms.VMs {
		log.Println(vm.Name)
	}
}

type VMs struct {
	VMs []VM `xml:"vm"`
}

type VM struct {
	Name string `xml:"name"`
}
