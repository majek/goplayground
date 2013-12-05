package lrucache

func (l *List) PushElementFront(e *Element) *Element {
	return l.insert(e, &l.root)
}

func (l *List) PushElementBack(e *Element) *Element {
	return l.insert(e, l.root.prev)
}

func (l *List) PopElementFront() *Element {
	el := l.Front()
	l.Remove(el)
	return el
}

func (l *List) PopFront() interface{} {
	el := l.Front()
	l.Remove(el)
	return el.Value
}

