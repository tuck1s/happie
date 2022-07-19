package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
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
			"Enclose IPv6 addresses with [] - e.g. [2a00:1450:400c:c0a::1b]:smtp" +
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

	conn, err := happie.New(proxy_addr_port, source_addr_port, dest_addr_port)
	if err != nil {
		log.Fatal(err)
	}

	var hdr []byte
	var version string
	if *proxy_protocol_v1 {
		hdr, err = conn.V1_Bytes()
		version = "1"
	} else {
		hdr, err = conn.V2_Bytes()
		version = "2"
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Sending header version %s\n%s\n", version, hex.Dump(hdr))
	reply := proxy_request(proxy_addr_port, hdr)
	fmt.Printf("Reply: %s\n", string(reply))
}
