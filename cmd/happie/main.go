package main

import (
	"bytes"
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

	hdr := happie.Connection(proxy_addr_port, source_addr_port, dest_addr_port)
	_ = hdr

	if *proxy_protocol_v1 {
		source_addr, source_port := happie.CheckedSplitHostPort(source_addr_port)
		dest_addr, dest_port := happie.CheckedSplitHostPort(dest_addr_port)
		hapv1 := fmt.Sprintf("PROXY TCP4 %s %s %d %d\r\n", source_addr, dest_addr, source_port, dest_port)
		fmt.Printf("Sending v1 header %s", hapv1)
		reply := proxy_request(proxy_addr_port, []byte(hapv1))
		fmt.Printf("Reply: %s\n", string(reply))

	} else {
		buf := new(bytes.Buffer)
		fmt.Printf("Sending v2 header\n%s\n", hex.Dump(buf.Bytes()))
		reply := proxy_request(proxy_addr_port, buf.Bytes())
		fmt.Printf("Reply: %s\n", string(reply))
	}
}
