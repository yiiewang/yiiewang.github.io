// per-slot 序列号版 MPMC 无锁队列：基础使用 + 压测。
//
// 队列实现与博客《两个生产者之后》路线三（07-1 §4.1）一致：
// Vyukov bounded MPMC——每个槽位一个序列号 seq，记录"本槽下一次可写时的绝对位置"，
// 生产者 CAS 认领 enqPos，消费者 CAS 认领 deqPos，发布/释放都落在 per-slot 状态机上。
//
// 运行（go.mod 在 example/ 下）：
//
//	cd example
//	go run ./20260907-1/per-slot                       # 基础使用 + 默认压测矩阵
//	go run ./20260907-1/per-slot -p 8 -c 8 -n 2000000  # 自定义压测参数
//	go run -race ./20260907-1/per-slot -n 20000        # 竞态检测（数据量压小）
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// -----------------------------------------------------------------------------
// 压测框架
// -----------------------------------------------------------------------------

// chanQueue 是带缓冲的 Go channel 基线：非阻塞收发 + 调用方自旋，
// 与 MPMC"满则立刻失败"的语义对齐。
type chanQueue struct {
	ch    chan entry
	abort *atomic.Bool
	full  atomic.Int64 // 发送落空次数
}

func (q *chanQueue) Push(e entry) bool {
	for {
		select {
		case q.ch <- e:
			return true
		default:
			q.full.Add(1)
			if q.abort.Load() {
				return false
			}
			runtime.Gosched()
		}
	}
}

func (q *chanQueue) Pop() (entry, bool) {
	select {
	case v := <-q.ch:
		return v, true
	default:
		return 0, false
	}
}

func (q *chanQueue) Spins() (int64, int64) { return q.full.Load(), 0 }

const seqBits = 32 // entry 编码：高 32 位生产者 id，低 32 位生产者内序号

func encode(pid int, seq int64) entry {
	return entry(uint64(pid)<<seqBits | uint64(seq))
}

func decode(v entry) (pid int, seq int64) {
	return int(uint64(v) >> seqBits), int64(uint64(v) & ((1 << seqBits) - 1))
}

type result struct {
	impl        string
	p, c        int
	n, capacity int64
	elapsed     time.Duration
	full        int64 // 队满自旋总次数
	empty       int64 // 队空自旋总次数
	err         error
}

func (r result) items() int64 { return int64(r.p) * r.n }

func (r result) ops() float64 {
	return float64(r.items()) / r.elapsed.Seconds()
}

type queueFactory func(abort *atomic.Bool) queue

// stress 跑一轮 P 生产者 × C 消费者压测：
// 生产者各发 n 条（值 = pid<<32|seq），消费者校验每条恰好一次、无丢失无重复；
// 单消费者时额外逐条校验生产者内 FIFO；统计吞吐与满/空自旋次数。
func stress(impl string, newQ queueFactory, p, c int, n, capacity int64, timeout time.Duration) result {
	res := result{impl: impl, p: p, c: c, n: n, capacity: capacity}
	var abort atomic.Bool
	q := newQ(&abort)
	total := int64(p) * n
	deadline := time.Now().Add(timeout)

	// flags[pid][seq]：CAS 0→1，失败即重复投递
	flags := make([][]int32, p)
	for i := range flags {
		flags[i] = make([]int32, n)
	}
	fullHits := make([]int64, p)  // 各生产者队满自旋次数（本地计数，热路径无原子开销）
	emptyHits := make([]int64, c) // 各消费者队空自旋次数

	var consumed int64
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error
	setErr := func(err error) {
		errMu.Lock()
		if firstErr == nil {
			firstErr = err
		}
		errMu.Unlock()
	}

	start := time.Now()
	for pid := 0; pid < p; pid++ {
		wg.Add(1)
		go func(pid int) {
			defer wg.Done()
			for seq := int64(0); seq < n; seq++ {
				for !q.Push(encode(pid, seq)) {
					fullHits[pid]++
					if abort.Load() {
						return
					}
					runtime.Gosched()
				}
			}
		}(pid)
	}
	for cid := 0; cid < c; cid++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()
			var lastSeq []int64 // 仅单消费者时逐条校验生产者内 FIFO
			if c == 1 {
				lastSeq = make([]int64, p)
				for i := range lastSeq {
					lastSeq[i] = -1 // 哨兵：尚未收到该生产者的任何消息
				}
			}
			for {
				if atomic.LoadInt64(&consumed) >= total {
					return
				}
				v, ok := q.Pop()
				if !ok {
					emptyHits[cid]++
					if time.Now().After(deadline) {
						abort.Store(true)
						setErr(fmt.Errorf("看门狗超时（%v）：仅消费 %d/%d 条，疑似丢数据", timeout, atomic.LoadInt64(&consumed), total))
						return
					}
					runtime.Gosched()
					continue
				}
				atomic.AddInt64(&consumed, 1)
				pid, seq := decode(v)
				if pid >= p || seq >= n {
					abort.Store(true)
					setErr(fmt.Errorf("收到非法值 %d（pid=%d seq=%d）", uint64(v), pid, seq))
					return
				}
				if !atomic.CompareAndSwapInt32(&flags[pid][seq], 0, 1) {
					abort.Store(true)
					setErr(fmt.Errorf("pid=%d seq=%d 被消费了两次（重复投递）", pid, seq))
					return
				}
				if lastSeq != nil {
					if seq != lastSeq[pid]+1 {
						abort.Store(true)
						setErr(fmt.Errorf("生产者 %d 的 FIFO 顺序破坏：期望 %d，实际 %d", pid, lastSeq[pid]+1, seq))
						return
					}
					lastSeq[pid] = seq
				}
			}
		}(cid)
	}
	wg.Wait()
	res.elapsed = time.Since(start)

	qFull, qEmpty := q.Spins()
	for _, v := range fullHits {
		res.full += v
	}
	for _, v := range emptyHits {
		res.empty += v
	}
	res.full += qFull
	res.empty += qEmpty

	if firstErr == nil {
		firstErr = checkComplete(flags, p, n)
	}
	res.err = firstErr
	return res
}

// checkComplete 校验每个 (pid, seq) 都被消费过恰好一次（无丢失）。
func checkComplete(flags [][]int32, p int, n int64) error {
	for pid := 0; pid < p; pid++ {
		for seq := int64(0); seq < n; seq++ {
			if atomic.LoadInt32(&flags[pid][seq]) == 0 {
				return fmt.Errorf("生产者 %d 的第 %d 条消息丢失", pid, seq)
			}
		}
	}
	return nil
}

func mpmcFactory(capacity int64) queueFactory {
	return func(*atomic.Bool) queue { return NewMPMC(capacity) }
}

func chanFactory(capacity int64) queueFactory {
	return func(abort *atomic.Bool) queue {
		return &chanQueue{ch: make(chan entry, roundPow2(capacity)), abort: abort}
	}
}

// -----------------------------------------------------------------------------
// 基础使用演示
// -----------------------------------------------------------------------------

func demoBasic() {
	fmt.Println("== 基础使用：单 goroutine 语义 ==")
	q := NewMPMC(4) // 容量自动向上取整到 2 的幂

	if _, ok := q.Pop(); ok {
		panic("空队列 Pop 不应成功")
	}
	fmt.Println("[1] 空队列 Pop       → ok=false ✓")

	for i := entry(0); i < 4; i++ {
		if !q.Push(i) {
			panic("未满队列 Push 不应失败")
		}
	}
	for i := entry(0); i < 4; i++ {
		if v, ok := q.Pop(); !ok || v != i {
			panic(fmt.Sprintf("FIFO 破坏：期望 %d，实际 %d（ok=%v）", i, v, ok))
		}
	}
	fmt.Println("[2] FIFO 顺序        → 0,1,2,3 依序出队 ✓")

	for i := entry(0); i < 4; i++ {
		if !q.Push(i) {
			panic("未满队列 Push 不应失败")
		}
	}
	if q.Push(99) {
		panic("满队列 Push 不应成功")
	}
	fmt.Println("[3] 满队列 Push      → ok=false（无副作用，不覆盖）✓")

	if _, ok := q.Pop(); !ok {
		panic("非空队列 Pop 应成功")
	}
	if !q.Push(100) {
		panic("腾出一格后 Push 应成功")
	}
	fmt.Println("[4] 腾出一格再 Push  → ok=true（环形复用）✓")
	fmt.Println()
}

func demoMulti(timeout time.Duration) {
	fmt.Println("== 基础使用：多生产者 × 多消费者 ==")
	// 2×1：单消费者下可逐条校验"生产者内 FIFO"
	if r := stress("MPMC", mpmcFactory(16), 2, 1, 2000, 16, timeout); r.err != nil {
		panic(r.err)
	}
	fmt.Println("[5] 2生产者 × 1消费者（各2000条）→ 恰好一次，且每个生产者的 FIFO 顺序逐条校验通过 ✓")
	// 2×2：多消费者校验恰好一次 + 全量覆盖
	if r := stress("MPMC", mpmcFactory(16), 2, 2, 2000, 16, timeout); r.err != nil {
		panic(r.err)
	}
	fmt.Println("[6] 2生产者 × 2消费者（各2000条）→ 每条恰好一次，无丢失无重复 ✓")
	fmt.Println()
}

// -----------------------------------------------------------------------------
// 报表输出
// -----------------------------------------------------------------------------

func printHeader() {
	fmt.Printf("%-9s %6s %8s %10s %10s %12s %12s %12s\n",
		"实现", "P×C", "容量", "条数", "耗时", "吞吐", "满自旋", "空自旋")
}

func (r result) row() string {
	return fmt.Sprintf("%-9s %6s %8d %10s %10s %12s %12d %12d",
		r.impl,
		fmt.Sprintf("%d×%d", r.p, r.c),
		r.capacity,
		humanCount(r.items()),
		r.elapsed.Round(time.Millisecond),
		humanOps(r.ops()),
		r.full,
		r.empty)
}

func humanCount(n int64) string {
	switch f := float64(n); {
	case f >= 1e9:
		return fmt.Sprintf("%.2fG", f/1e9)
	case f >= 1e6:
		return fmt.Sprintf("%.2fM", f/1e6)
	case f >= 1e3:
		return fmt.Sprintf("%.1fk", f/1e3)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func humanOps(ops float64) string {
	switch {
	case ops >= 1e9:
		return fmt.Sprintf("%.2fG/s", ops/1e9)
	case ops >= 1e6:
		return fmt.Sprintf("%.2fM/s", ops/1e6)
	case ops >= 1e3:
		return fmt.Sprintf("%.1fk/s", ops/1e3)
	default:
		return fmt.Sprintf("%.0f/s", ops)
	}
}

// -----------------------------------------------------------------------------
// main
// -----------------------------------------------------------------------------

func main() {
	pFlag := flag.Int("p", 0, "压测生产者数（与 -c 同时指定时只跑这一组）")
	cFlag := flag.Int("c", 0, "压测消费者数（与 -p 同时指定时只跑这一组）")
	nFlag := flag.Int64("n", 200000, "每个生产者发送的条数")
	capFlag := flag.Int64("cap", 1024, "队列容量（自动向上取整到 2 的幂）")
	timeoutFlag := flag.Duration("timeout", 30*time.Second, "单轮压测看门狗超时")
	flag.Parse()

	fmt.Printf("GOMAXPROCS=%d  CPU=%d\n", runtime.GOMAXPROCS(0), runtime.NumCPU())

	demoBasic()
	demoMulti(*timeoutFlag)

	pairs := [][2]int{{1, 1}, {2, 2}, {4, 4}, {8, 8}, {8, 1}, {1, 8}}
	if *pFlag > 0 && *cFlag > 0 {
		pairs = [][2]int{{*pFlag, *cFlag}}
	}

	fmt.Println("== 压测：每条消息校验恰好一次；满/空立即失败 + Gosched 自旋重试（MPMC 与 channel 语义对齐） ==")
	fmt.Println("   （满自旋/空自旋 = 队列满/空时调用方重试的次数，反映竞争烈度）")
	fmt.Println()
	printHeader()
	failed := false
	for _, pc := range pairs {
		p, c := pc[0], pc[1]
		r1 := stress("MPMC", mpmcFactory(*capFlag), p, c, *nFlag, *capFlag, *timeoutFlag)
		r2 := stress("channel", chanFactory(*capFlag), p, c, *nFlag, *capFlag, *timeoutFlag)
		for _, r := range []result{r1, r2} {
			fmt.Println(r.row())
			if r.err != nil {
				failed = true
				fmt.Printf("    ✗ %v\n", r.err)
			}
		}
		fmt.Println()
	}
	if failed {
		os.Exit(1)
	}
	fmt.Println("全部通过：每条消息恰好被消费一次，无丢失、无重复、FIFO 语义保持。")
}
