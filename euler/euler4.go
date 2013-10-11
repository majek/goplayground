package main

import (
	"fmt"
)

func palindrome(v int) bool {
	s := fmt.Sprintf("%d", v)
	l := len(s)
	for i := 0; i < l/2; i++ {
		if s[i] != s[l-1-i] {
			return false
		}
	}
	return true
}

func main() {
	m := 0
	for i := 999; i > 99; i-- {
		for j := 999; j > 99; j-- {
			v := i * j
			if v > m {
				if palindrome(v) {
					m = v
				}
			} else {
				// v aint gonna grow.
				break
			}
		}
	}
	fmt.Printf("%d\n", m)
}
