package main

// Disruptor：装配体。零件（§1-§3）在这里接成依赖图（§3.3）
type Disruptor struct {
	rb   *RingBuffer
	seqr Sequencer
	stop chan struct{}
}

func NewDisruptor(sizePow2 int, seqr Sequencer) *Disruptor {
	return &Disruptor{rb: NewRingBuffer(sizePow2), seqr: seqr, stop: make(chan struct{})}
}

// HandleWith：注册一个并行消费者，直接消费 ring buffer
func (d *Disruptor) HandleWith(wait WaitStrategy, on func(*Event, int64, bool)) *BatchProcessor {
	p := &BatchProcessor{barrier: &Barrier{seqr: d.seqr, wait: wait}, rb: d.rb, onEvent: on}
	d.seqr.AddGating(&p.seq)
	go p.Run(d.stop)
	return p
}

// After：注册链式消费者——prev 全部推进到位，才轮到它
func (d *Disruptor) After(prev []*BatchProcessor, wait WaitStrategy,
	on func(*Event, int64, bool)) *BatchProcessor {
	deps := make([]*Sequence, len(prev))
	for i, p := range prev {
		deps[i] = &p.seq
	}
	p := &BatchProcessor{barrier: &Barrier{seqr: d.seqr, deps: deps, wait: wait},
		rb: d.rb, onEvent: on}
	d.seqr.AddGating(&p.seq)
	go p.Run(d.stop)
	return p
}

// Publish：生产者侧两步——先写槽位，后发布（release 语义，07.md 的"发布 + 观察"）
func (d *Disruptor) Publish(ev *Event) {
	seq := d.seqr.Next()
	*d.rb.slot(seq) = *ev
	d.seqr.Publish(seq)
}
