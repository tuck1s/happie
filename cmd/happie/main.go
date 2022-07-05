package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math"
	"net"
	"os"

	"github.com/tuck1s/happie"
)

func checkedWrite(con net.Conn, b []byte) {
	_, err := con.Write(b)
	if err != nil {
		log.Fatal(err)
	}
}

func checkedRead(con net.Conn) []byte {
	reply := make([]byte, 1024)
	_, err := con.Read(reply)
	if err != nil {
		log.Fatal(err)
	}
	return reply
}

func checkedSplitHostPort(addr string) (string, int) {
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
	host, p_int := checkedSplitHostPort(addr)
	host_ip4 := net.ParseIP(host).To4()
	if host_ip4 == nil {
		log.Fatalf("IP address %s invalid.", host)
	}
	host_u32 := binary.BigEndian.Uint32(host_ip4)
	return host_u32, uint16(p_int)
}

func proxy_request(proxy_addr_port string, req []byte) []byte {
	con, err := net.Dial("tcp", proxy_addr_port)
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()
	checkedWrite(con, req)
	return checkedRead(con)
}

func main() {
	proxy_protocol_v1 := flag.Bool("v1", false, "Use PROXY protocol v1 header")
	flag.Usage = func() {
		const helpText = "HAProxy tester\n" +
			"Usage of %[1]s:\n" +
			"    %[1]s [FLAGS] proxy:port source:port dest:port\n" +
			"    ports can be names (e.g. smtp) or numbers (e.g. 25).\n\n" +
			"proxy:port\tProxy listening for your request, e.g. 127.0.0.1:5000.\n\n" +
			"source:port\tAddress on the proxy used for onward conection.\n" +
			"\t\tMust be an address hosted by the proxy itself, otherwise the request will fail.\n" +
			"\t\tSet to :0 to have the proxy choose an ephemeral port.\n\n" +
			"dest:port\tThe service the proxy should connect to. e.g. 64.233.167.27:smtp (Google mail server).\n\n" +
			"FLAGS:\n"
		fmt.Fprintf(flag.CommandLine.Output(), helpText, os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() < 3 {
		flag.Usage()
		os.Exit(1)
	}
	proxy_addr_port := flag.Arg(0)
	source_addr_port := flag.Arg(1)
	dest_addr_port := flag.Arg(2)

	if *proxy_protocol_v1 {
		source_addr, source_port := checkedSplitHostPort(source_addr_port)
		dest_addr, dest_port := checkedSplitHostPort(dest_addr_port)
		hapv1 := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n", source_addr, dest_addr, source_port, dest_port)
		fmt.Printf("Sending v1 header %s", hapv1)
		reply := proxy_request(proxy_addr_port, []byte(hapv1))
		fmt.Printf("Reply: %s\n", string(reply))

	} else {
		var hapv2 happie.HAproxy_header_v2
		// PROXY protocol v2 signature - 12 fixed bytes at the start
		hapv2.sig1 = binary.BigEndian.Uint32([]byte("\r\n\r\n"))
		hapv2.sig2 = binary.BigEndian.Uint32([]byte("\x00\r\nQ"))
		hapv2.sig3 = binary.BigEndian.Uint32([]byte("UIT\n"))
		hapv2.version_command = 0x21   // 2 = Version 2, 1 = request comes from a proxy
		hapv2.addr_family_proto = 0x11 // 1 = AF_INET, 1 = STREAM (TCP)

		hapv2.ipv4_addr.source, hapv2.ipv4_addr.dest_port = addr_port_to_u32_u16(source_addr_port)
		hapv2.ipv4_addr.dest, hapv2.ipv4_addr.dest_port = addr_port_to_u32_u16(dest_addr_port)

		// Address field containst source + dest + source_port + dest_port
		hapv2.addr_length = uint16(binary.Size(hapv2.ipv4_addr))
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, hapv2)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Sending v2 header\n%s\n", hex.Dump(buf.Bytes()))
		reply := proxy_request(proxy_addr_port, buf.Bytes())
		fmt.Printf("Reply: %s\n", string(reply))
	}
}
