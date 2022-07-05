package happie

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"net"
)

type HAproxy_header_v2 struct {
	sig1, sig2, sig3  uint32 // 12 byte signature at start
	version_command   uint8
	addr_family_proto uint8
	addr_length       uint16
	ipv4_addr         struct {
		source      uint32
		dest        uint32
		source_port uint16
		dest_port   uint16
	}
}

type HAproxy_conn struct {
	proxy_addr_port  string
	source_addr_port string
	dest_addr_port   string
}

func Connection(proxy_addr_port, source_addr_port, dest_addr_port string) *HAproxy_conn {
	hdr := HAproxy_conn{
		proxy_addr_port:  proxy_addr_port,
		source_addr_port: source_addr_port,
		dest_addr_port:   dest_addr_port,
	}
	return &hdr
}

func CheckedSplitHostPort(addr string) (string, int) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}

	p_int, err := net.LookupPort("tcp4", port)
	if err != nil {
		log.Fatal(err)
	}

	// Check the port number is within valid range
	if p_int < 0 || p_int > math.MaxUint16 {
		log.Fatalf("Port number must be 0 .. %d.", math.MaxUint16)
	}

	return host, p_int
}

func addr_port_to_u32_u16(addr string) (uint32, uint16) {
	host, p_int := CheckedSplitHostPort(addr)
	host_ip4 := net.ParseIP(host).To4()
	if host_ip4 == nil {
		log.Fatalf("IP address %s invalid.", host)
	}
	host_u32 := binary.BigEndian.Uint32(host_ip4)
	return host_u32, uint16(p_int)
}

func (hdr *HAproxy_conn) HeaderBytes() {
	var hapv2 HAproxy_header_v2
	// PROXY protocol v2 signature - 12 fixed bytes at the start
	hapv2.sig1 = binary.BigEndian.Uint32([]byte("\r\n\r\n"))
	hapv2.sig2 = binary.BigEndian.Uint32([]byte("\x00\r\nQ"))
	hapv2.sig3 = binary.BigEndian.Uint32([]byte("UIT\n"))
	hapv2.version_command = 0x21   // 2 = Version 2, 1 = request comes from a proxy
	hapv2.addr_family_proto = 0x11 // 1 = AF_INET, 1 = STREAM (TCP)

	hapv2.ipv4_addr.source, hapv2.ipv4_addr.dest_port = addr_port_to_u32_u16(hdr.source_addr_port)
	hapv2.ipv4_addr.dest, hapv2.ipv4_addr.dest_port = addr_port_to_u32_u16(hdr.dest_addr_port)

	// Address field containst source + dest + source_port + dest_port
	hapv2.addr_length = uint16(binary.Size(hapv2.ipv4_addr))
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, hapv2)
	if err != nil {
		log.Fatal(err)
	}

}
