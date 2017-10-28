package host

import (
	"sync"

	"github.com/imjoey/go-ovirt"
	"github.com/prometheus/common/log"
)

var (
	cacheMutex = sync.Mutex{}
	nameCache  = make(map[string]string)
)

func Follow(h *ovirtsdk.Host, conn *ovirtsdk.Connection) (*ovirtsdk.Host, error) {
	x, err := conn.FollowLink(h)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	h = x.(*ovirtsdk.Host)
	return h, nil
}

func Name(h *ovirtsdk.Host, conn *ovirtsdk.Connection) string {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if n, found := nameCache[h.MustId()]; found {
		return n
	}

	h, err := Follow(h, conn)
	if err != nil {
		log.Error(err)
		return ""
	}

	nameCache[h.MustId()] = h.MustName()
	return h.MustName()
}
