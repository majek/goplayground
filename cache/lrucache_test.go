package cache

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	t.Parallel()
	b := NewLRUCache(3)
	if b.Get("a") != nil {
		t.Error("")
	}

	now := time.Now()
	b.Set("b", "vb", now.Add(time.Duration(2*time.Second)))
	b.Set("a", "va", now.Add(time.Duration(1*time.Second)))
	b.Set("c", "vc", now.Add(time.Duration(3*time.Second)))

	if b.Get("a") != "va" {
		t.Error("")
	}
	if b.Get("b") != "vb" {
		t.Error("")
	}
	if b.Get("c") != "vc" {
		t.Error("")
	}

	b.Set("d", "vd", now.Add(time.Duration(4*time.Second)))
	if b.Get("a") != nil {
		t.Error("Expecting element A to be evicted")
	}

	b.Set("e", "ve", now.Add(time.Duration(-4*time.Second)))
	if b.Get("b") != nil {
		t.Error("Expecting element B to be evicted")
	}

	b.Set("f", "vf", now.Add(time.Duration(5*time.Second)))
	if b.Get("e") != nil {
		t.Error("Expecting element E to be evicted")
	}

	if b.Get("c") == nil {
		t.Error("Expecting element C to not be evicted")
	}
	n := now.Add(time.Duration(10 * time.Second))
	b.SetNow("g", "vg", now.Add(time.Duration(5*time.Second)), &n)
	if b.Get("c") != nil {
		t.Error("Expecting element C to be evicted")
	}

	if b.Len() != 3 {
		t.Error("Expecting different length")
	}
	b.Del("miss")
	b.Del("g")
	if b.Len() != 2 {
		t.Error("Expecting different length")
	}

	b.Clear()
	if b.Len() != 0 {
		t.Error("Expecting different length")
	}

	now = time.Now()
	b.Set("b", "vb", now.Add(time.Duration(2*time.Second)))
	b.Set("a", "va", now.Add(time.Duration(1*time.Second)))
	b.Set("d", "vd", now.Add(time.Duration(4*time.Second)))
	b.Set("c", "vc", now.Add(time.Duration(3*time.Second)))

	if b.Get("b") != nil {
		t.Error("Expecting miss")
	}

	b.GetQuiet("miss")
	if b.GetQuiet("a") != "va" {
		t.Error("Expecting hit")
	}

	b.Set("e", "ve", now.Add(time.Duration(5*time.Second)))
	if b.Get("a") != nil {
		t.Error("Expecting miss")
	}

}

func TestPanicOnNil(t *testing.T) {
	t.Parallel()
	b := NewLRUCache(3)

	recovered := false
	defer func() {
		if r := recover(); r != nil {
			recovered = true
		}
	}()

	b.Set("a", nil, time.Now())

	if !recovered {
		t.Error("Expected Set to raise panic")
	}
}

func TestExtra(t *testing.T) {
	t.Parallel()
	b := NewLRUCache(3)
	if b.Get("a") != nil {
		t.Error("")
	}

	now := time.Now()
	b.Set("b", "vb", now)
	b.Set("a", "va", now)
	b.Set("c", "vc", now.Add(time.Duration(3*time.Second)))

	if b.Get("a") != "va" {
		t.Error("expecting value")
	}

	if b.GetNotStale("a") != nil {
		t.Error("not expecting value")
	}
	if b.GetNotStale("miss") != nil {
		t.Error("not expecting value")
	}
	if b.GetNotStale("c") != "vc" {
		t.Error("expecting hit")
	}

	if b.Len() != 2 {
		t.Error("Expecting different length")
	}
	if b.Expire() != 1 {
		t.Error("Expecting different length")
	}
}

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(65 + rand.Intn(90-65))
	}
	return string(bytes)
}

func createFilledBucket() *LRUCache {
	b := NewLRUCache(1000)
	expire := time.Now().Add(time.Duration(4))
	for i := 0; i < 1000; i++ {
		b.Set(randomString(2), "value", expire)
	}
	return b
}

func TestConcurrentGet(t *testing.T) {
	t.Parallel()
	b := createFilledBucket()

	done := make(chan bool)
	worker := func() {
		for i := 0; i < 10000; i++ {
			b.Get(randomString(2))
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func TestConcurrentSet(t *testing.T) {
	t.Parallel()
	b := createFilledBucket()

	done := make(chan bool)
	worker := func() {
		expire := time.Now().Add(time.Duration(4 * time.Second))
		for i := 0; i < 10000; i++ {
			b.Set(randomString(2), "value", expire)
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func BenchmarkConcurrentGet(bb *testing.B) {
	b := createFilledBucket()

	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			b.Get(randomString(2))
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}

func BenchmarkConcurrentSet(bb *testing.B) {
	b := createFilledBucket()

	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			expire := time.Now().Add(time.Duration(4 * time.Second))
			b.Set(randomString(2), "v", expire)
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}
