package main

import "sync/atomic"

type Sequence struct {
	_ [7]int64
	v int64
	_ [7]int64
}

func (s *Sequence) Get() int64  { return atomic.LoadInt64(&s.v) }
func (s *Sequence) Set(v int64) { atomic.StoreInt64(&s.v, v) }
func (s *Sequence) CAS(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&s.v, old, new)
}

// minSeq：起点和一串序号取最小——"最慢的消费者在哪"只此一个问法
func minSeq(start int64, seqs ...*Sequence) int64 {
	m := start
	for _, s := range seqs {
		if v := s.Get(); v < m {
			m = v
		}
	}
	return m
}
