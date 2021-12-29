package p2p

import (
	"net"

	tmsync "github.com/tendermint/tendermint/libs/sync"
)

// ConnSet is a lookup table for connections and all their ips.
type ConnSet interface {
	Has(ID, net.Conn) bool
	HasIP(net.IP) bool
	Set(ID, net.Conn, []net.IP)
	Remove(ID, net.Conn)
	RemoveAddr(ID, net.Addr)
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

func (cs *connSet) Has(id ID, c net.Conn) bool {
	cs.RLock()
	defer cs.RUnlock()

	_, ok := cs.conns[string(id)]

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

func (cs *connSet) Remove(id ID, c net.Conn) {
	cs.Lock()
	defer cs.Unlock()
	delete(cs.conns, string(id))
}

func (cs *connSet) RemoveAddr(id ID, addr net.Addr) {
	cs.Lock()
	defer cs.Unlock()

	delete(cs.conns, string(id))
}

func (cs *connSet) Set(id ID, c net.Conn, ips []net.IP) {
	cs.Lock()
	defer cs.Unlock()

	cs.conns[string(id)] = connSetItem{
		conn: c,
		ips:  ips,
	}
}
