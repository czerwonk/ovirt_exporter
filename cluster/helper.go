package cluster

import (
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/common/log"
)

func Follow(cl *ovirtsdk4.Cluster, conn *ovirtsdk4.Connection) (*ovirtsdk4.Cluster, error) {
	x, err := conn.FollowLink(cl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	cl = x.(*ovirtsdk4.Cluster)
	return cl, nil
}
