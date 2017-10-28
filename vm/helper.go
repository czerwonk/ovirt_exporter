package vm

import (
	"github.com/imjoey/go-ovirt"
)

func followSnapShots(d *ovirtsdk.SnapshotSlice, conn *ovirtsdk.Connection) (*ovirtsdk.SnapshotSlice, error) {
	x, err := conn.FollowLink(d)
	if err != nil {
		return nil, err
	}

	d = x.(*ovirtsdk.SnapshotSlice)
	return d, nil
}
