package vm

import "time"

type Snapshots struct {
	Snapshot []Snapshot `xml:"snapshot"`
}

type Snapshot struct {
	Id                 string    `xml:"id,attr"`
	Description        string    `xml:"description"`
	Date               time.Time `xml:"date"`
	PersistMemorystate bool      `xml:"persist_memorystate"`
	Status             string    `json:"snapshot_status"`
	Type               string    `json:"snapshot_type"`
}
