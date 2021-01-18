// Modified for Tendermint
// Originally Copyright (c) 2013-2014 Conformal Systems LLC.
// https://github.com/conformal/btcd/blob/master/LICENSE

package p2p

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/tendermint/tendermint/config"
	log2 "github.com/tendermint/tendermint/libs/log"
	tls2 "github.com/tendermint/tendermint/p2p/tls"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)


var (
	Config *config.Config
	loger log2.Logger
	//node_store *NodeStore
)

func SetP2PConfig(conf *config.Config) {
	Config = conf
}

func SetLogger(logger log2.Logger) {
	loger = logger
}

//func SetStore(s *NodeStore) {
//	node_store = s
//}

// NetAddress defines information about a peer on the network
// including its ID, IP address, and port.
type NetAddress struct {
	ID   ID     `json:"id"`
	IP   net.IP `json:"ip"`
	Port uint16 `json:"port"`

	Host  string  `json:"host"`

	// TODO:
	// Name string `json:"name"` // optional DNS name

	// memoize .String()
	str string
}

// IDAddressString returns id@hostPort. It strips the leading
// protocol from protocolHostPort if it exists.
func IDAddressString(id ID, protocolHostPort string) string {
	hostPort := removeProtocolIfDefined(protocolHostPort)
	return fmt.Sprintf("%s@%s", id, hostPort)
}

// NewNetAddress returns a new NetAddress using the provided TCP
// address. When testing, other net.Addr (except TCP) will result in
// using 0.0.0.0:0. When normal run, other net.Addr (except TCP) will
// panic. Panics if ID is invalid.
// TODO: socks proxies?
func NewNetAddress(id ID, addr net.Addr) *NetAddress {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		if flag.Lookup("test.v") == nil { // normal run
			panic(fmt.Sprintf("Only TCPAddrs are supported. Got: %v", addr))
		} else { // in testing
			netAddr := NewNetAddressIPPort(net.IP("127.0.0.1"), 0)
			netAddr.ID = id
			return netAddr
		}
	}

	if err := validateID(id); err != nil {
		panic(fmt.Sprintf("Invalid ID %v: %v (addr: %v)", id, err, addr))
	}

	ip := tcpAddr.IP
	port := uint16(tcpAddr.Port)
	na := NewNetAddressIPPort(ip, port)
	na.ID = id
	//host := ""
	//na.Host = host
	return na
}

/////http request.
/////get server_name of ip.
//func queryServerName(ip net.IP) string {
//	if Config.P2P.TLSOption {
//		return node_store.Get(ip.String())
//	}else {
//		return ip.String()
//	}
//}

// NewNetAddressString returns a new NetAddress using the provided address in
// the form of "ID@IP:Port".
// Also resolves the host if host is not an IP.
// Errors are of type ErrNetAddressXxx where Xxx is in (NoID, Invalid, Lookup)
func NewNetAddressString(addr string) (*NetAddress, error) {
	addrWithoutProtocol := removeProtocolIfDefined(addr)
	spl := strings.Split(addrWithoutProtocol, "@")
	if len(spl) != 2 {
		return nil, ErrNetAddressNoID{addr}
	}

	// get ID
	if err := validateID(ID(spl[0])); err != nil {
		return nil, ErrNetAddressInvalid{addrWithoutProtocol, err}
	}
	var id ID
	id, addrWithoutProtocol = ID(spl[0]), spl[1]

	// get host and port
	host, portStr, err := net.SplitHostPort(addrWithoutProtocol)
	if err != nil {
		return nil, ErrNetAddressInvalid{addrWithoutProtocol, err}
	}
	if len(host) == 0 {
		return nil, ErrNetAddressInvalid{
			addrWithoutProtocol,
			errors.New("host is empty")}
	}

	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, ErrNetAddressLookup{host, err}
		}
		ip = ips[0]
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, ErrNetAddressInvalid{portStr, err}
	}

	na := NewNetAddressIPPort(ip, uint16(port))
	na.ID = id
	na.Host = host
	return na, nil
}

// NewNetAddressStrings returns an array of NetAddress'es build using
// the provided strings.
func NewNetAddressStrings(addrs []string) ([]*NetAddress, []error) {
	netAddrs := make([]*NetAddress, 0)
	errs := make([]error, 0)
	for _, addr := range addrs {
		netAddr, err := NewNetAddressString(addr)
		if err != nil {
			errs = append(errs, err)
		} else {
			netAddrs = append(netAddrs, netAddr)
		}
	}
	return netAddrs, errs
}

// NewNetAddressIPPort returns a new NetAddress using the provided IP
// and port number.
func NewNetAddressIPPort(ip net.IP, port uint16) *NetAddress {
	return &NetAddress{
		IP:   ip,
		Port: port,
	}
}

// Equals reports whether na and other are the same addresses,
// including their ID, IP, and Port.
func (na *NetAddress) Equals(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		return na.String() == o.String()
	}
	return false
}

// Same returns true is na has the same non-empty ID or DialString as other.
func (na *NetAddress) Same(other interface{}) bool {
	if o, ok := other.(*NetAddress); ok {
		if na.DialString() == o.DialString() {
			return true
		}
		if na.ID != "" && na.ID == o.ID {
			return true
		}
	}
	return false
}

// String representation: <ID>@<IP>:<PORT>
func (na *NetAddress) String() string {
	if na == nil {
		return "<nil-NetAddress>"
	}
	if na.str == "" {
		addrStr := na.DialString()
		if na.ID != "" {
			addrStr = IDAddressString(na.ID, addrStr)
		}
		na.str = addrStr
	}
	return na.str
}

func (na *NetAddress) DialString() string {
	if na == nil {
		return "<nil-NetAddress>"
	}
	return net.JoinHostPort(
		na.IP.String(),
		strconv.FormatUint(uint64(na.Port), 10),
	)
}

// Dial calls net.Dial on the address.
func (na *NetAddress) Dial() (net.Conn, error) {
	conn, err := net.Dial("tcp", na.DialString())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// DialTimeout calls net.DialTimeout on the address.
func (na *NetAddress) DialTimeout2(timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", na.DialString(), timeout)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (na *NetAddress) DialTimeout(timeout time.Duration) (net.Conn, error) {
	loger.Info("TLS Option:", "tls", Config.P2P.TLSOption)
	loger.Info("TLS Config:", "local_bind_address_ip",Config.TLSConfig.BindAddressIP,
		"local_bind_port", Config.TLSConfig.BindAddressPort,
		"remote_host", Config.TLSConfig.RemoteAddressHOST,
		"remote_port", Config.TLSConfig.RemoteAddressPort,
		"remote_server_name", Config.TLSConfig.RemoteServerName,
		"remote_insecure_skip_verify", Config.TLSConfig.RemoteTLSInsecureSkipVerify,
	)
	if Config.P2P.TLSOption {
		Config.TLSConfig.Mutex.Lock()
		loger.Info("get netAddress", "host", na.Host, "port", na.Port)
		tls2.SetRemoteNOdeAddress(na.Host, int(na.Port))
		Config.TLSConfig.Mutex.Unlock()
		localHost := fmt.Sprintf("%s:%s", Config.TLSConfig.BindAddressIP, strconv.FormatUint(uint64(Config.TLSConfig.BindAddressPort), 10))
		loger.Info("dial host", "host", localHost)
		conn, err := net.DialTimeout("tcp", localHost, timeout)
		if err != nil {
			loger.Error("dial error", "err", err.Error())
			return nil, err
		}
		return conn, nil
	}else {
		return na.DialTimeout2(timeout)
	}
}

// Routable returns true if the address is routable.
func (na *NetAddress) Routable() bool {
	if err := na.Valid(); err != nil {
		return false
	}
	// TODO(oga) bitcoind doesn't include RFC3849 here, but should we?
	return !(na.RFC1918() || na.RFC3927() || na.RFC4862() ||
		na.RFC4193() || na.RFC4843() || na.Local())
}

// For IPv4 these are either a 0 or all bits set address. For IPv6 a zero
// address or one that matches the RFC3849 documentation address format.
func (na *NetAddress) Valid() error {
	if err := validateID(na.ID); err != nil {
		return errors.Wrap(err, "invalid ID")
	}

	if na.IP == nil {
		return errors.New("no IP")
	}
	if na.IP.IsUnspecified() || na.RFC3849() || na.IP.Equal(net.IPv4bcast) {
		return errors.New("invalid IP")
	}
	return nil
}

// HasID returns true if the address has an ID.
// NOTE: It does not check whether the ID is valid or not.
func (na *NetAddress) HasID() bool {
	return string(na.ID) != ""
}

// Local returns true if it is a local address.
func (na *NetAddress) Local() bool {
	return na.IP.IsLoopback() || zero4.Contains(na.IP)
}

// ReachabilityTo checks whenever o can be reached from na.
func (na *NetAddress) ReachabilityTo(o *NetAddress) int {
	const (
		Unreachable = 0
		Default     = iota
		Teredo
		Ipv6Weak
		Ipv4
		Ipv6Strong
	)
	switch {
	case !na.Routable():
		return Unreachable
	case na.RFC4380():
		switch {
		case !o.Routable():
			return Default
		case o.RFC4380():
			return Teredo
		case o.IP.To4() != nil:
			return Ipv4
		default: // ipv6
			return Ipv6Weak
		}
	case na.IP.To4() != nil:
		if o.Routable() && o.IP.To4() != nil {
			return Ipv4
		}
		return Default
	default: /* ipv6 */
		var tunnelled bool
		// Is our v6 is tunnelled?
		if o.RFC3964() || o.RFC6052() || o.RFC6145() {
			tunnelled = true
		}
		switch {
		case !o.Routable():
			return Default
		case o.RFC4380():
			return Teredo
		case o.IP.To4() != nil:
			return Ipv4
		case tunnelled:
			// only prioritise ipv6 if we aren't tunnelling it.
			return Ipv6Weak
		}
		return Ipv6Strong
	}
}

// RFC1918: IPv4 Private networks (10.0.0.0/8, 192.168.0.0/16, 172.16.0.0/12)
// RFC3849: IPv6 Documentation address  (2001:0DB8::/32)
// RFC3927: IPv4 Autoconfig (169.254.0.0/16)
// RFC3964: IPv6 6to4 (2002::/16)
// RFC4193: IPv6 unique local (FC00::/7)
// RFC4380: IPv6 Teredo tunneling (2001::/32)
// RFC4843: IPv6 ORCHID: (2001:10::/28)
// RFC4862: IPv6 Autoconfig (FE80::/64)
// RFC6052: IPv6 well known prefix (64:FF9B::/96)
// RFC6145: IPv6 IPv4 translated address ::FFFF:0:0:0/96
var rfc1918_10 = net.IPNet{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 32)}
var rfc1918_192 = net.IPNet{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 32)}
var rfc1918_172 = net.IPNet{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 32)}
var rfc3849 = net.IPNet{IP: net.ParseIP("2001:0DB8::"), Mask: net.CIDRMask(32, 128)}
var rfc3927 = net.IPNet{IP: net.ParseIP("169.254.0.0"), Mask: net.CIDRMask(16, 32)}
var rfc3964 = net.IPNet{IP: net.ParseIP("2002::"), Mask: net.CIDRMask(16, 128)}
var rfc4193 = net.IPNet{IP: net.ParseIP("FC00::"), Mask: net.CIDRMask(7, 128)}
var rfc4380 = net.IPNet{IP: net.ParseIP("2001::"), Mask: net.CIDRMask(32, 128)}
var rfc4843 = net.IPNet{IP: net.ParseIP("2001:10::"), Mask: net.CIDRMask(28, 128)}
var rfc4862 = net.IPNet{IP: net.ParseIP("FE80::"), Mask: net.CIDRMask(64, 128)}
var rfc6052 = net.IPNet{IP: net.ParseIP("64:FF9B::"), Mask: net.CIDRMask(96, 128)}
var rfc6145 = net.IPNet{IP: net.ParseIP("::FFFF:0:0:0"), Mask: net.CIDRMask(96, 128)}
var zero4 = net.IPNet{IP: net.ParseIP("0.0.0.0"), Mask: net.CIDRMask(8, 32)}

func (na *NetAddress) RFC1918() bool {
	return rfc1918_10.Contains(na.IP) ||
		rfc1918_192.Contains(na.IP) ||
		rfc1918_172.Contains(na.IP)
}
func (na *NetAddress) RFC3849() bool { return rfc3849.Contains(na.IP) }
func (na *NetAddress) RFC3927() bool { return rfc3927.Contains(na.IP) }
func (na *NetAddress) RFC3964() bool { return rfc3964.Contains(na.IP) }
func (na *NetAddress) RFC4193() bool { return rfc4193.Contains(na.IP) }
func (na *NetAddress) RFC4380() bool { return rfc4380.Contains(na.IP) }
func (na *NetAddress) RFC4843() bool { return rfc4843.Contains(na.IP) }
func (na *NetAddress) RFC4862() bool { return rfc4862.Contains(na.IP) }
func (na *NetAddress) RFC6052() bool { return rfc6052.Contains(na.IP) }
func (na *NetAddress) RFC6145() bool { return rfc6145.Contains(na.IP) }

func removeProtocolIfDefined(addr string) string {
	if strings.Contains(addr, "://") {
		return strings.Split(addr, "://")[1]
	}
	return addr

}

func validateID(id ID) error {
	if len(id) == 0 {
		return errors.New("no ID")
	}
	idBytes, err := hex.DecodeString(string(id))
	if err != nil {
		return err
	}
	if len(idBytes) != IDByteLength {
		return fmt.Errorf("invalid hex length - got %d, expected %d", len(idBytes), IDByteLength)
	}
	return nil
}
