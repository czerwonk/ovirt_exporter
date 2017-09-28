package datacenter

import (
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/common/log"
)

func Follow(d *ovirtsdk4.DataCenter, conn *ovirtsdk4.Connection) (*ovirtsdk4.DataCenter, error) {
	x, err := conn.FollowLink(d)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	d = x.(*ovirtsdk4.DataCenter)
	return d, nil
}
