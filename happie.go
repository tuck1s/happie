package happie

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
