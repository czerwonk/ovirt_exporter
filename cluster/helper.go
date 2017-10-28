package cluster

import (
	"sync"

	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/common/log"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

func Follow(cl *ovirtsdk.Cluster, conn *ovirtsdk.Connection) (*ovirtsdk.Cluster, error) {
	x, err := conn.FollowLink(cl)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	cl = x.(*ovirtsdk.Cluster)
	return cl, nil
}

func Name(c *ovirtsdk.Cluster, conn *ovirtsdk.Connection) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[c.MustId()]; found {
		return n
	}

	c, err := Follow(c, conn)
	if err != nil {
		log.Error(err)
		return ""
	}

	nameCache[c.MustId()] = c.MustName()
	return c.MustName()
}
