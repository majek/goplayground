package main

import (
	"fmt"
	"net"
	"os"
)

func unpackDns(msg []byte) (domain string, id uint16, ips []net.IP) {
	d := new(dnsMsg)
	if !d.Unpack(msg) {
		// fmt.Fprintf(os.Stderr, "dns error (unpacking)\n")
		return
	}

	id = d.id

	if len(d.question) < 1 {
		// fmt.Fprintf(os.Stderr, "dns error (wrong question section)\n")
		return
	}

	domain = d.question[0].Name
	if len(domain) < 1 {
		// fmt.Fprintf(os.Stderr, "dns error (wrong domain in question)\n")
		return
	}

	_, addrs, err := answer(domain, "server", d, dnsTypeA)
	if err == nil {
		ips = convertRR_A(addrs)
	}
	return
}

func packDns(domain string, id uint16) []byte {

	out := new(dnsMsg)
	out.id = id
	out.recursion_desired = true
	out.question = []dnsQuestion{
		{domain, dnsTypeA, dnsClassINET},
	}

	msg, ok := out.Pack()
	if !ok {
		fmt.Fprintf(os.Stderr, "can't pack domain %s\n", domain)
		os.Exit(1)
	}
	return msg
}
