package p2p

import (
	"net"

	tmsync "github.com/tendermint/tendermint/libs/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(string, net.Conn) bool
	HasIP(net.IP) bool
	Set(string, net.Conn, []net.IP)
	Remove(string, net.Conn)
	RemoveAddr(string, net.Addr)
}

type connSetItem struct {
	conn net.Conn
	ips  []net.IP
}

type connSet struct {
	tmsync.RWMutex

	conns map[string]connSetItem
}

// NewConnSet returns a ConnSet implementation.
func NewConnSet() ConnSet {
	return &connSet{
		conns: map[string]connSetItem{},
	}
}

func (cs *connSet) Has(host string, c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[host]

	return ok
}

func (cs *connSet) HasIP(ip net.IP) bool {
	cs.RLock()
	defer cs.RUnlock()

	for _, c := range cs.conns {
		for _, known := range c.ips {
			if known.Equal(ip) {
				return true
			}
		}
	}

	return false
}

func (cs *connSet) Remove(host string, c net.Conn) {
	cs.Lock()
	defer cs.Unlock()
	if len(host) < 1 {
		host = c.RemoteAddr().String()
	}
	delete(cs.conns, host)
}

func (cs *connSet) RemoveAddr(host string, addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()
	if len(host) < 1 {
		host = addr.String()
	}
	delete(cs.conns, host)
}

func (cs *connSet) Set(host string, c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()
	if len(host) < 1 {
		host = c.RemoteAddr().String()
	}

	cs.conns[host] = connSetItem{
		conn: c,
		ips:  ips,
	}

}
