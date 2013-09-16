package bitmap_test

import (
	. "github.com/majek/bitmap"
	"testing"
)

func TestBitmap(t *testing.T) {
	b := NewBitmap()
	for i := 0; i < 100; i++ {
		if b.Get(i) != false {
			t.Fail()
		}
	}

	b.Set(1, true)

	if b.Get(0) != false {
		t.Fail()
	}
	if b.Get(1) != true {
		t.Fail()
	}
	if b.Get(2) != false {
		t.Fail()
	}

	i := 0
	for v := range b.Iter() {
		if i == 1 {
			if v != true {
				t.Fail()
			}
		} else {
			if v != false {
				t.Fail()
			}
		}
		i = i + 1
	}
	if i != 2 {
		t.Fail()
	}
}
