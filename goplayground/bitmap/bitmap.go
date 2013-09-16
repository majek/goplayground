package bitmap

type Bitmap struct {
	a         []byte
	max_index int
}

func NewBitmapLength(length_hint uint) *Bitmap {
	return &Bitmap{make([]byte, (length_hint/8)+1), 0}
}

func NewBitmap() *Bitmap {
	return NewBitmapLength(0)
}

func (b *Bitmap) Get(n int) bool {
	idx, off := n/8, n%8
	if idx >= len(b.a) {
		return false
	}
	return b.a[idx]&(1<<uint(off)) != 0
}

func (b *Bitmap) Set(n int, v bool) {
	if b.max_index <= n {
		b.max_index = n + 1
	}
	idx, off := n/8, n%8
	if idx >= len(b.a) {
		new_a := make([]byte, idx*2)
		copy(new_a, b.a)
		b.a = new_a
	}
	if v == true {
		b.a[idx] = b.a[idx] | 1<<uint(off)
	} else {
		b.a[idx] = b.a[idx] &^ 1 << uint(off)
	}
}

func (b *Bitmap) Iter() <-chan bool {
	c := make(chan bool)
	go func() {
		for i := 0; i < b.max_index; i++ {
			idx, off := i/8, i%8
			c <- b.a[idx]&(1<<uint(off)) != 0
		}
		close(c)
	}()
	return c
}
