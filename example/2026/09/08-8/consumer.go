package main

// Barrier：消费者的"能追到哪"。先等前置进度，再被收紧到连续已发布界
type Barrier struct {
	seqr Sequencer
	deps []*Sequence // 前置消费者的序号：链式依赖的"闸门"
	wait WaitStrategy
}

// WaitFor：先等"能追到哪"，再被 availableBuffer 收紧到"连续能追到哪"
func (b *Barrier) WaitFor(seq int64) int64 {
	avail := b.wait.WaitFor(seq, b.seqr.Cursor(), b.deps)
	if avail < seq {
		return avail // 队列空
	}
	return b.seqr.HighestPublished(seq, avail)
}

// BatchProcessor：消费者驱动器。业务逻辑一行不掺，全在装配时注册的 onEvent 里
type BatchProcessor struct {
	seq     Sequence // 自己的消费进度：唯一写者是自己
	barrier *Barrier
	rb      *RingBuffer
	onEvent func(ev *Event, seq int64, endOfBatch bool) // 业务回调：装配时注册
}

func (p *BatchProcessor) Run(stop <-chan struct{}) {
	next := p.seq.Get() + 1
	for {
		select {
		case <-stop:
			return
		default:
		}
		avail := p.barrier.WaitFor(next)
		if avail >= next {
			for i := next; i <= avail; i++ {
				p.onEvent(p.rb.slot(i), i, i == avail) // endOfBatch 收口
			}
			next = avail + 1
		}
		p.seq.Set(avail) // 发布进度：生产者据此判断槽位可否复用
	}
}
