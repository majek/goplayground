package main

import (
	"encoding/binary"
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"bytes"
)

type mappingType struct {
	ipv4 [32+1]map[uint32]string
	ipv6 [64+1]map[uint64]string
}

func (M *mappingType) init() {
	for i := 0; i < 32+1; i++ {
		M.ipv4[i] = make(map[uint32]string)
	}
	for i := 0; i < 64+1; i++ {
		M.ipv6[i] = make(map[uint64]string)
	}
}
func (M *mappingType) match(ip net.IP) (bool, string) {
	buf := bytes.NewReader(ip)
	if len(ip) == 4 {
		var numIp uint32
		binary.Read(buf, binary.BigEndian, &numIp)

		for i := 32; i > 0; i-- {
			ma := numIp & (((1 << uint(i)) - 1) << (32-uint(i)))
			v, ok := M.ipv4[i][uint32(ma)]
			if ok {
				return true, v
			}
		}
	} else if len(ip) == 16 {
		var numIp uint64
		binary.Read(buf, binary.BigEndian, &numIp)

		for i := 64; i > 0; i-- {
			ma := numIp & (((1 << uint(i)) - 1) << (64-uint(i)))
			v, ok := M.ipv6[i][ma]
			if ok {
				return true, v
			}
		}
	} else {
		panic("")
	}
	return false, ""
}

func (M *mappingType)load(fname string, delimiters string) {
	fd, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer fd.Close()

	split := prepareSplit(delimiters)

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

		parts := split(line)

		_, ipnet, err := net.ParseCIDR(parts[0])
		if err != nil || len(parts) < 2 {
			fmt.Fprintf(os.Stderr, "line %d can't parse network %#v: %s\n", i,
				line, err)
			os.Exit(1)
		}
		buf := bytes.NewReader(ipnet.IP)
		mask, _ := ipnet.Mask.Size()
		if len(ipnet.IP) == 4 {
			var numIp uint32
			binary.Read(buf, binary.BigEndian, &numIp)
			M.ipv4[mask][numIp] = parts[1]
		} else if len(ipnet.IP) == 8 {
			var numIp uint64
			binary.Read(buf, binary.BigEndian, &numIp)
			M.ipv6[mask][numIp] = parts[1]
		} else {
			panic("")
		}
	}
	return
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
			"\"ip2as\" prints out a line only when it contains an " +
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
	var M mappingType
	M.init()
	M.load(flag.Arg(0), " \t")

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
		ip := net.ParseIP(parts[0])
		if ip == nil {
			continue
		}
		x := ip.To4()
		if x != nil {
			ip = x
		}
		ok, as := M.match(ip)
		if ok {
			fmt.Printf("%s\n", as)
		} else {
			fmt.Printf("UNKNOWN\n")
		}
	}
}
