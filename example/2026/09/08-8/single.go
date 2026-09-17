package main

import "runtime"

type SingleProducerSequencer struct {
	bufSize int64
	cursor  Sequence // 发布边界：唯一由 Publish 推进
	gating  []*Sequence
	claim   int64 // 认领边界：只有生产者 goroutine 碰，普通变量即可
	cached  int64 // 缓存的最慢消费者：避免每次跨核读
}

func NewSingleProducer(bufSize int64) *SingleProducerSequencer {
	return &SingleProducerSequencer{bufSize: bufSize}
}

// Cursor：发布边界。消费者看这里判断"发布到哪了"
func (p *SingleProducerSequencer) Cursor() *Sequence { return &p.cursor }

// AddGating：登记消费者进度名单。必须先于第一次 Publish——运行期变更就是数据竞争
func (p *SingleProducerSequencer) AddGating(s ...*Sequence) { p.gating = append(p.gating, s...) }

func (p *SingleProducerSequencer) Next() int64 {
	next := p.claim + 1
	if wrap := next - p.bufSize; wrap > p.cached {
		p.cached = minSeq(p.cursor.Get(), p.gating...) // 缓存失效才跨核
		for wrap > p.cached {
			runtime.Gosched() // 满：等最慢消费者让位
			p.cached = minSeq(p.cursor.Get(), p.gating...)
		}
	}
	p.claim = next
	return next
}

func (p *SingleProducerSequencer) Publish(seq int64) {
	p.cursor.Set(seq)
}

func (p *SingleProducerSequencer) HighestPublished(lower, upper int64) int64 {
	return upper // 单生产者发布必然连续：cursor 到哪，哪就是连续就绪
}
