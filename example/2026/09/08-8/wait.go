package main

import "runtime"

// WaitStrategy：可插拔的"等"。完整档位谱系（自旋→让步→park→阻塞唤醒）见 08-2
type WaitStrategy interface {
	WaitFor(seq int64, cursor *Sequence, deps []*Sequence) int64
}

// BusySpin：纯自旋，延迟最低、烧核。LMAX 的交易管线用它
type BusySpin struct{}

func (BusySpin) WaitFor(seq int64, cursor *Sequence, deps []*Sequence) int64 {
	for {
		if avail := minSeq(cursor.Get(), deps...); avail >= seq {
			return avail
		}
	}
}

// Yielding：自旋 N 次让一步。后台管线用，省核
type Yielding struct{}

func (Yielding) WaitFor(seq int64, cursor *Sequence, deps []*Sequence) int64 {
	for i := 0; ; i++ {
		if avail := minSeq(cursor.Get(), deps...); avail >= seq {
			return avail
		}
		if i&0x3F == 0x3F {
			runtime.Gosched() // 每 64 次 Gosched 一轮
		}
	}
}
