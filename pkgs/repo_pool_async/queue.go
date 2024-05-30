package repo_pool_async

type qNode struct {
	data any
	next *qNode
}
type Queue struct {
	f *qNode
	b *qNode
	l int
}

func (q *Queue) En(data any) {
	n := qNode{data: data}
	if q.f == nil {
		q.f, q.b = &n, &n
	} else {
		oldB := q.b
		q.b = &n
		oldB.next = &n
	}

	q.l++
}

func (q *Queue) De() any {
	if q.f != nil {
		oldF := q.f
		q.f = oldF.next
		q.l--
		return oldF.data
	}

	return nil
}

func (q *Queue) Empty() bool {
	return q.f == nil
}

func (q *Queue) Size() int {
	return q.l
}
