package vm

type Vms struct {
	Vm []Vm `xml:"vm"`
}

type Vm struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name"`
}
