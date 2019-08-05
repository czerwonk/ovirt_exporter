package vm

import "time"

// Snapshots is a collection of snapshots
type Snapshots struct {
	Snapshot []Snapshot `xml:"snapshot"`
}

// Snapshot repesents the snapshot resource
type Snapshot struct {
	ID                 string    `xml:"id,attr"`
	Description        string    `xml:"description"`
	Date               time.Time `xml:"date"`
	PersistMemorystate bool      `xml:"persist_memorystate"`
	Status             string    `json:"snapshot_status"`
	Type               string    `json:"snapshot_type"`
}
