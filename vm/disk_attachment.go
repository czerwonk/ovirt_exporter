// SPDX-License-Identifier: MIT

package vm

import "github.com/czerwonk/ovirt_exporter/disk"

// DiskAttachments is a collection of diskattachments
type DiskAttachments struct {
	Attachment []DiskAttachment `xml:"disk_attachment"`
}

// DiskAttachment represents the diskattachment resource
type DiskAttachment struct {
	LogicalName string    `xml:"logical_name"`
	Disk        disk.Disk `xml:"disk"`
}
