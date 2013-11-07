package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

func do_read_domains(domains chan<- string, domainSlotAvailable <-chan bool) {
	in := bufio.NewReader(os.Stdin)

	for _ = range domainSlotAvailable {

		input, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "read(stdin): %s\n", err)
			os.Exit(1)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		domain := input + "."

		domains <- domain
	}
	close(domains)
}

var sendingDelay time.Duration
var retryDelay time.Duration

var concurrency int
var dnsServer string
var packetsPerSecond int
var retryTime string
var verbose bool

func init() {
	flag.StringVar(&dnsServer, "server", "8.8.8.8:53",
		"DNS server address (ip:port)")
	flag.IntVar(&concurrency, "concurrency", 5000,
		"Internal buffer")
	flag.IntVar(&packetsPerSecond, "pps", 120,
		"Send up to PPS DNS queries per second")
	flag.StringVar(&retryTime, "retry", "1s",
		"Resend unanswered query after RETRY")
	flag.BoolVar(&verbose, "v", false,
		"Verbose logging")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, strings.Join([]string{
			"\"resolve\" mass resolve DNS A records for domains names read from stdin.",
			"",
			"Usage: resolve [option ...]",
			"",
		}, "\n"))
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 0 {
		flag.Usage()
		os.Exit(1)
	}

	sendingDelay = time.Duration(1000000000/packetsPerSecond) * time.Nanosecond
	var err error
	retryDelay, err = time.ParseDuration(retryTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't parse duration %s\n", retryTime)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Server: %s, sending delay: %s (%d pps), retry delay: %s\n",
		dnsServer, sendingDelay, packetsPerSecond, retryDelay)

	domains := make(chan string, concurrency)
	domainSlotAvailable := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		domainSlotAvailable <- true
	}

	go do_read_domains(domains, domainSlotAvailable)

	c, err := net.Dial("udp", dnsServer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bind(udp, %s): %s\n", dnsServer, err)
		os.Exit(1)
	}

	// Used as a queue. Make sure it has plenty of storage available.
	timeoutRegister := make(chan *domainRecord, concurrency*1000)
	timeoutExpired := make(chan *domainRecord)

	resolved := make(chan *domainAnswer, concurrency)
	tryResolving := make(chan *domainRecord, concurrency)

	go do_timeouter(timeoutRegister, timeoutExpired)

	go do_send(c, tryResolving)
	go do_receive(c, resolved)

	t0 := time.Now()
	domainsCount, avgTries := do_map_guard(domains, domainSlotAvailable,
		timeoutRegister, timeoutExpired,
		tryResolving, resolved)
	td := time.Now().Sub(t0)
	fmt.Fprintf(os.Stderr, "Resolved %d domains in %.3fs. Average retries %.3f. Domains per second: %.3f\n",
		domainsCount,
		td.Seconds(),
		avgTries,
		float64(domainsCount)/td.Seconds())
}

type domainRecord struct {
	id      uint16
	domain  string
	timeout time.Time
	resend  int
}

type domainAnswer struct {
	id     uint16
	domain string
	ips    []net.IP
}

func do_map_guard(domains <-chan string,
	domainSlotAvailable chan<- bool,
	timeoutRegister chan<- *domainRecord,
	timeoutExpired <-chan *domainRecord,
	tryResolving chan<- *domainRecord,
	resolved <-chan *domainAnswer) (int, float64) {

	m := make(map[uint16]*domainRecord)

	done := false

	sumTries := 0
	domainCount := 0

	for done == false || len(m) > 0 {
		select {
		case domain := <-domains:
			if domain == "" {
				domains = make(chan string)
				done = true
				break
			}
			var id uint16
			for {
				id = uint16(rand.Int())
				if id != 0 && m[id] == nil {
					break
				}
			}
			dr := &domainRecord{id, domain, time.Now(), 1}
			m[id] = dr
			if verbose {
				fmt.Fprintf(os.Stderr, "0x%04x resolving %s\n", id, domain)
			}
			timeoutRegister <- dr
			tryResolving <- dr

		case dr := <-timeoutExpired:
			if m[dr.id] == dr {
				dr.resend += 1
				dr.timeout = time.Now()
				if verbose {
					fmt.Fprintf(os.Stderr, "0x%04x resend (try:%d) %s\n", dr.id,
						dr.resend, dr.domain)
				}
				timeoutRegister <- dr
				tryResolving <- dr
			}

		case da := <-resolved:
			if m[da.id] != nil {
				dr := m[da.id]
				if dr.domain != da.domain {
					if verbose {
						fmt.Fprintf(os.Stderr, "0x%04x error, unrecognized domain: %s != %s\n",
							da.id, dr.domain, da.domain)
					}
					break
				}

				if verbose {
					fmt.Fprintf(os.Stderr, "0x%04x resolved %s\n",
						dr.id, dr.domain)
				}

				s := make([]string, 0, 16)
				for _, ip := range da.ips {
					s = append(s, ip.String())
				}
				sort.Sort(sort.StringSlice(s))

				// without trailing dot
				domain := dr.domain[:len(dr.domain)-1]
				fmt.Printf("%s, %s\n", domain, strings.Join(s, " "))

				sumTries += dr.resend
				domainCount += 1

				delete(m, dr.id)
				domainSlotAvailable <- true
			}
		}
	}
	return domainCount, float64(sumTries) / float64(domainCount)
}

func do_timeouter(timeoutRegister <-chan *domainRecord,
	timeoutExpired chan<- *domainRecord) {
	for {
		dr := <-timeoutRegister
		t := dr.timeout.Add(retryDelay)
		now := time.Now()
		if t.Sub(now) > 0 {
			delta := t.Sub(now)
			time.Sleep(delta)
		}
		timeoutExpired <- dr
	}
}

func do_send(c net.Conn, tryResolving <-chan *domainRecord) {
	for {
		dr := <-tryResolving

		msg := packDns(dr.domain, dr.id)

		_, err := c.Write(msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "write(udp): %s\n", err)
			os.Exit(1)
		}
		time.Sleep(sendingDelay)
	}
}

func do_receive(c net.Conn, resolved chan<- *domainAnswer) {
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		domain, id, ips := unpackDns(buf[:n])
		resolved <- &domainAnswer{id, domain, ips}
	}
}
