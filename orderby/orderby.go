package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sort"
)

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

func loadFile(fname string, ch chan<- string ) {
	fd, err := os.Open(fname)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer fd.Close()

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
		ch <- line
	}
	close(ch)
}


var delimiters string
var fname string

func init() {
	flag.StringVar(&delimiters, "delimiters", ", \t;[]",
		"delimiters")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, strings.Join([]string{
			"\"orderby\" reads the stdin and prints it out",
			"in order given by the order file",
			"",
			"",
			"Usage: orderby [-delimiters=...] ORDERFILE",
			"",
		}, "\n"))
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	ch := make(chan string, 1024)
	split := prepareSplit(delimiters)
	go loadFile(flag.Arg(0), ch)

	m := make(map[string][]string)

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
		key := split(line)[0]
		m[key] = append(m[key], line)
	}

	for key := range ch {
		for _, line := range m[key] {
			fmt.Printf("%s\n", line)
		}
		delete(m,key)
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		for _, line := range m[key] {
			fmt.Printf("%s\n", line)
		}
		delete(m,key)
	}
}
