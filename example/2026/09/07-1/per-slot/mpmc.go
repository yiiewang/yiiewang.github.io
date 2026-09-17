package main

import "sync/atomic"

// queue 统一 MPMC 与 channel 基线的接口：满/空都立即返回 false，
// 由调用方 Gosched 自旋重试——两种实现的等待语义对齐，吞吐才有可比性。
type queue interface {
	Push(e entry) bool
	Pop() (entry, bool)
	// Spins 返回队列内部自旋计数（channel 的 select 落空在内部发生，需要单独记）
	Spins() (full, empty int64)
}

type entry uint64

// cell 前两个 8 字节是状态机与数据，尾部补齐成独立缓存行，
// 避免相邻槽位的写互相打掉对方的缓存行（false sharing）。
type cell struct {
	seq  int64 // 状态机：== pos 可写，== pos+1 可读
	data entry
	_    [56]byte // 补齐到独立 64B 缓存行（按 entry 大小调整）
}

type MPMC struct {
	buf    []cell // 容量为 2 的幂
	mask   int64
	enqPos int64 // 生产者 CAS 认领
	deqPos int64 // 消费者 CAS 认领
}

func NewMPMC(capacity int64) *MPMC {
	capacity = roundPow2(capacity)
	q := &MPMC{buf: make([]cell, capacity), mask: capacity - 1}
	for i := range q.buf {
		q.buf[i].seq = int64(i) // 初始：槽 i 属于第一圈的 pos=i
	}
	return q
}

// roundPow2 把容量向上取整到最近的 2 的幂（如 1000→1024、5→8），已是 2 的幂则原样返回。
//
// 队列用 pos&mask 定位槽位，它等价于 pos%capacity 的前提是 capacity 为 2 的幂
// （此时 mask=capacity-1 二进制全 1）；否则按位与会把读写指到错误槽位，
// per-slot 状态机"发布 +1 / 释放 +cap，一圈恰好 cap 个槽位"的递推也随之失效。
// 顺带的收益：热路径上的除法取模换成单条 AND 指令。
func roundPow2(n int64) int64 {
	if n < 1 {
		n = 1 // 防御：0 / 负容量钳到最小值 1
	}
	if n&(n-1) == 0 {
		// 2 的幂判定：2 的幂二进制只有一个 1，减 1 后该位变 0、低位全变 1，
		// 按位与必为 0（如 8&7=0b1000&0b0111=0；非 2 的幂如 6&5=0b100≠0）
		return n
	}
	p := int64(1)
	for p < n {
		p <<= 1 // 从 1 开始不断翻倍，直到 ≥ n
	}
	return p
}

func (q *MPMC) Push(e entry) bool {
	pos := atomic.LoadInt64(&q.enqPos)
	for {
		c := &q.buf[pos&q.mask]
		switch seq := atomic.LoadInt64(&c.seq); {
		case seq == pos: // 轮到我：本圈此槽可写
			if atomic.CompareAndSwapInt64(&q.enqPos, pos, pos+1) {
				c.data = e                       // 先写数据
				atomic.StoreInt64(&c.seq, pos+1) // 后发布（release）
				return true
			}
		case seq < pos: // 上一圈的数据还没被消费 → 满
			return false
		default: // 别人先认领了，刷新重试
			pos = atomic.LoadInt64(&q.enqPos)
		}
	}
}

func (q *MPMC) Pop() (entry, bool) {
	pos := atomic.LoadInt64(&q.deqPos)
	for {
		c := &q.buf[pos&q.mask]
		switch seq := atomic.LoadInt64(&c.seq); {
		case seq == pos+1: // 本圈数据就位
			if atomic.CompareAndSwapInt64(&q.deqPos, pos, pos+1) {
				e := c.data                             // 先读数据
				atomic.StoreInt64(&c.seq, pos+q.mask+1) // 后释放（= pos+cap，下圈可写）
				return e, true
			}
		case seq < pos+1: // 本圈还没发布 → 空
			return 0, false
		default: // 别人先认领了，刷新重试
			pos = atomic.LoadInt64(&q.deqPos)
		}
	}
}

// Spins 返回队列内部自旋计数。MPMC 的自旋发生在调用方（压测框架统计），恒为 0。
func (q *MPMC) Spins() (full, empty int64) { return 0, 0 }
