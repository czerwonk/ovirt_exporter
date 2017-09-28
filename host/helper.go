package host

import (
	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/common/log"
)

func Follow(h *ovirtsdk4.Host, conn *ovirtsdk4.Connection) (*ovirtsdk4.Host, error) {
	x, err := conn.FollowLink(h)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	h = x.(*ovirtsdk4.Host)
	return h, nil
}
