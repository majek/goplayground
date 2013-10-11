package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func loadNetworksFile(fname string) []*net.IPNet {
	fd, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer fd.Close()

	ipnets := make([]*net.IPNet, 0, 200)

	in := bufio.NewReader(fd)
	for i := 0; ; i++ {
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		_, ipnet, err := net.ParseCIDR(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "line %d can't parse network %#v: %s\n", i,
				line, err)
			os.Exit(1)
		}
		ipnets = append(ipnets, ipnet)
	}
	return ipnets
}

func prepareSplit(delimiters string) func(string) []string {
	m := make(map[rune]bool)
	for _, c := range delimiters {
		m[c] = true
	}
	fun := func(line string) []string {
		p := make([]string, 0, 4)
		prev := 0
		for i, r := range line {
			if m[r] != false {
				p = append(p, line[prev:i])
				prev = i + 1
			}
		}
		return append(p, line[prev:])
	}
	return fun
}

var delimiters string
var fname string

func init() {
	flag.StringVar(&delimiters, "delimiters", ", \t;[]",
		"delimiters used to extract IP addresses from stdin")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, strings.Join([]string{
			"\"grepnet\" prints out a line only when it contains an " +
				"IP address from one of the specified networks.",
			"",
			"NETWORKFILE is a file that contains a list of matching " +
				"networks, one in a line. For exampe: " +
				"\"192.168.0.0/24\" or \"127.0.0.1/8\"",
			"",
			"Usage: grepnet [-delimiters=...] NETWORKFILE",
			"",
		}, "\n"))
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	split := prepareSplit(delimiters)
	networks := loadNetworksFile(flag.Arg(0))

	in := bufio.NewReader(os.Stdin)
	for {
		line, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if len(line) > 0 {
			// Truncate trailing new line
			line = line[:len(line)-1]
		}
		parts := split(line)
	LineLoop:
		for _, p := range parts {
			ip := net.ParseIP(p)
			if ip == nil {
				continue
			}
			for _, n := range networks {
				if n.Contains(ip) {
					fmt.Printf("%s\n", line)
					break LineLoop
				}
			}
		}
	}
}
