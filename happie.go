package happie

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"strings"
)

type HAproxy_conn struct {
	proxy,
	dest,
	source netip.AddrPort
}

// Split a string address and port into an AddrPort type, checking for errors
// and resolving named services to numeric port numbers
func split_ip_port(addr_port string) (netip.AddrPort, error) {
	a, p, err := net.SplitHostPort(addr_port)
	if err != nil {
		return netip.AddrPort{}, err
	}
	// parse the addresses
	addr, err := netip.ParseAddr(a)
	if err != nil {
		return netip.AddrPort{}, err
	}
	// resolve any named services (e.g. "smtp" to numeric port numbers)
	p_int, err := net.LookupPort("tcp4", p)
	if err != nil {
		return netip.AddrPort{}, err
	}
	if p_int < 0 || p_int > math.MaxUint16 {
		return netip.AddrPort{}, fmt.Errorf("Port number must be 0 .. %d.", math.MaxUint16)
	}
	return netip.AddrPortFrom(addr, uint16(p_int)), nil
}

// Register a new PROXY Protocol (https://www.haproxy.org/download/2.6/doc/proxy-protocol.txt) association, comprising:
//  The address:port to reach the proxy service.
//  source address:port - that the proxy should use for its outward connection.
//      This address must be hosted on the proxy.
//		You can specify port=0 to have the proxy choose an ephemeral port.
//  destination address:port - remote service that the proxy should connect to.
//  Addresses should be valid IPv4.
func New(proxy_addr_port, source_addr_port, dest_addr_port string) (*HAproxy_conn, error) {
	pa, err := split_ip_port(proxy_addr_port)
	if err != nil {
		return nil, err
	}
	sa, err := split_ip_port(source_addr_port)
	if err != nil {
		return nil, err
	}
	da, err := split_ip_port(dest_addr_port)
	if err != nil {
		return nil, err
	}
	conn := HAproxy_conn{
		proxy:  pa,
		source: sa,
		dest:   da,
	}
	return &conn, nil
}

// Return the PROXY connection header in version 1 (text) format, as a byte slice
// Currently this conversion is hardcoded for ipv4 source and dest addresses, and TCP connection type
// TODO: support IPv6 addresses and UDP connection types
func (c *HAproxy_conn) V1_Bytes() ([]byte, error) {
	hapv1 := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n",
		c.source.Addr().String(), c.dest.Addr().String(), c.source.Port(), c.dest.Port())
	// Conversion problems are signalled in the returned string - see https://pkg.go.dev/fmt#hdr-Format_errors
	if strings.Contains(hapv1, "%!") {
		return nil, errors.New(hapv1)
	} else {
		return []byte(hapv1), nil
	}
}

// Return the PROXY connection header in version 2 (binary) format, as a byte slice
// Currently this conversion is hardcoded for ipv4 source and dest addresses, and TCP connection type
// TODO: support IPv6 addresses and UDP connection types
func (c *HAproxy_conn) V2_Bytes() ([]byte, error) {
	hapv2_source, err := c.source.Addr().MarshalBinary()
	if err != nil {
		return nil, err
	}
	hapv2_dest, err := c.dest.Addr().MarshalBinary()
	if err != nil {
		return nil, err
	}

	header := new(bytes.Buffer)
	// PROXY protocol v2 signature - 12 fixed bytes at the start
	_, err = header.Write([]byte("\r\n\r\n\x00\r\nQUIT\n"))
	if err != nil {
		return nil, err
	}
	// TODO: need to set up family / proto according to address types.
	// I think that both addresses may need to be the SAME type (v4 / v6) - need to check that https://www.haproxy.org/download/1.8/doc/proxy-protocol.txt
	// 2 = Version 2, 1 = request comes from a proxy
	// 1 = AF_INET, 	1 = STREAM (TCP)
	_, err = header.Write([]byte("\x21\x11"))

	addrs := new(bytes.Buffer)
	// Address field containst source + dest + source_port + dest_port
	_, err = addrs.Write(hapv2_source)
	if err != nil {
		return nil, err
	}
	_, err = addrs.Write(hapv2_dest)
	if err != nil {
		return nil, err
	}

	err = binary.Write(addrs, binary.BigEndian, c.source.Port())
	if err != nil {
		return nil, err
	}

	err = binary.Write(addrs, binary.BigEndian, c.dest.Port())
	if err != nil {
		return nil, err
	}

	// Put the length of the composite address fields, then the addresses, into the header
	len := uint16(addrs.Len())
	err = binary.Write(header, binary.BigEndian, len)
	if err != nil {
		return nil, err
	}
	_, err = addrs.WriteTo(header)
	if err != nil {
		return nil, err
	}

	return header.Bytes(), nil
}
