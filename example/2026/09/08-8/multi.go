package main

import (
	"math/bits"
	"runtime"
	"sync/atomic"
)

type MultiProducerSequencer struct {
	bufSize   int64
	mask      int64
	shift     uint     // log2(bufSize)：seq >> shift = 第几圈
	cursor    Sequence // 认领边界：CAS 竞争点
	available []int32  // 每格"第几圈已发布"，-1 = 从未
	gating    []*Sequence
}

func NewMultiProducer(sizePow2 int) *MultiProducerSequencer {
	available := make([]int32, sizePow2)
	for i := range available {
		available[i] = -1 // -1 = 从未发布：上一圈残留的旧标记天然对不上本圈回合数
	}
	return &MultiProducerSequencer{
		bufSize:   int64(sizePow2),
		mask:      int64(sizePow2 - 1),
		shift:     uint(bits.TrailingZeros(uint(sizePow2))),
		available: available,
	}
}

// Cursor：认领边界（CAS 竞争点）。下游先拿它当追高上限，再被 HighestPublished 收紧到连续已发布界
func (p *MultiProducerSequencer) Cursor() *Sequence { return &p.cursor }

func (p *MultiProducerSequencer) AddGating(s ...*Sequence) { p.gating = append(p.gating, s...) }

// Next：CAS 认领一个槽位（演示版每次重读 gating，生产版加 2.1 的 cached）
func (p *MultiProducerSequencer) Next() int64 {
	for {
		cur := p.cursor.Get()
		next := cur + 1
		if wrap := next - p.bufSize; wrap > minSeq(cur, p.gating...) {
			runtime.Gosched()
			continue
		}
		if p.cursor.CAS(cur, next) {
			return cur // 认领成功，槽位号 = cur
		}
	}
}

// Publish：认领 ≠ 发布。数据写进槽位之后，才标记"本圈就绪"
func (p *MultiProducerSequencer) Publish(seq int64) {
	atomic.StoreInt32(&p.available[seq&p.mask], int32(seq>>p.shift))
}

// HighestPublished：把下游推进界从"认领界"收紧到"连续已发布界"
func (p *MultiProducerSequencer) HighestPublished(lower, upper int64) int64 {
	for seq := lower; seq <= upper; seq++ {
		if atomic.LoadInt32(&p.available[seq&p.mask]) != int32(seq>>p.shift) {
			return seq - 1 // 缺口（认领了没发布）：停在这里，FIFO 不跳过
		}
	}
	return upper
}
