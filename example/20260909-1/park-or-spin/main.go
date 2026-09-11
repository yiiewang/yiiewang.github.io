// park-or-spin：验证《队列满了，到底该自旋、让出还是睡觉》一文的实验。
//
//   -exp=stall   实验 1：SPSC 队列满之后，消费者多久能被调度？
//                三种等待策略：纯自旋 / Gosched 让出 / Sleep 睡眠
//   -exp=ladder  实验 2：各种"等一下"的单次成本阶梯（原子 / Gosched / channel park）
//
// 用法：
//   go run ./20260909-1/park-or-spin -exp=stall  -procs=1
//   go run ./20260909-1/park-or-spin -exp=stall  -procs=2
//   go run ./20260909-1/park-or-spin -exp=ladder -procs=1
//   go run ./20260909-1/park-or-spin -exp=ladder -procs=2
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"
)

// ---------- 与博客 07.md 相同的 SPSC 环形队列 ----------

type entry struct{ v int64 }

type SPSC struct {
	buf  []entry
	head int64 // 生产者独占写
	tail int64 // 消费者独占写
}

func newSPSC(capacity int) *SPSC { return &SPSC{buf: make([]entry, capacity)} }

func (q *SPSC) Push(e entry) bool {
	head := q.head
	if head-atomic.LoadInt64(&q.tail) >= int64(len(q.buf)) {
		return false // 满
	}
	q.buf[head&(int64(len(q.buf))-1)] = e
	atomic.StoreInt64(&q.head, head+1) // release
	return true
}

func (q *SPSC) Pop() (entry, bool) {
	tail := q.tail
	if tail == atomic.LoadInt64(&q.head) { // acquire
		return entry{}, false // 空
	}
	e := q.buf[tail&(int64(len(q.buf)-1))]
	atomic.StoreInt64(&q.tail, tail+1) // 释放槽位
	return e, true
}

// ---------- 实验 1：队列满之后，消费者多久能跑起来？ ----------

// 生产者全速灌 n 个元素，满时按策略等待；消费者逐个弹出并记录时间戳。
// 返回：首个元素延迟、相邻弹出最大间隔（被饿得最狠的一次）、总耗时、>1ms 的间隔个数。
func expStall(wait int, n, capacity int, sleepFor time.Duration) (firstLag, maxGap, total time.Duration, bigGaps int) {
	q := newSPSC(capacity)
	stamps := make([]time.Time, n)
	done := make(chan struct{})

	go func() { // 消费者
		cnt := 0
		for cnt < n {
			if e, ok := q.Pop(); ok {
				stamps[e.v] = time.Now()
				cnt++
			}
		}
		close(done)
	}()

	start := time.Now()
	for i := 0; i < n; i++ { // 生产者 = MergedQ.Run 的转发者
		for !q.Push(entry{v: int64(i)}) {
			switch wait {
			case 0: // 纯自旋：不让出
			case 1: // 自旋 + Gosched：让出一次再试
				runtime.Gosched()
			case 2: // 睡眠：真 park / 唤醒
				time.Sleep(sleepFor)
			}
		}
	}
	<-done
	total = time.Since(start)

	firstLag = stamps[0].Sub(start)
	for i := 1; i < n; i++ {
		gap := stamps[i].Sub(stamps[i-1])
		if gap > maxGap {
			maxGap = gap
		}
		if gap > time.Millisecond {
			bigGaps++
		}
	}
	return firstLag, maxGap, total, bigGaps
}

func runStall(n, capacity int, sleepFor time.Duration, rounds int, only int) {
	names := []string{"双端纯自旋", "双端自旋+Gosched", "双端自旋+Sleep"}
	for wait := 0; wait < 3; wait++ {
		if only >= 0 && wait != only {
			continue
		}
		fmt.Printf("\n--- 等待策略：%s ---\n", names[wait])
		for r := 0; r < rounds; r++ {
			firstLag, maxGap, total, big := expStall(wait, n, capacity, sleepFor)
			fmt.Printf("第%d轮  首元素延迟 %12v  最大弹出间隔 %12v  总耗时 %12v  (>1ms 间隔 %d 个)\n",
				r+1, firstLag, maxGap, total, big)
		}
	}
}

// ---------- 实验 2："等一下"的成本阶梯 ----------

// 原子标志往返：两边纯自旋，零调度参与（需要 ≥2 个 P）
func benchAtomicRound(m int) time.Duration {
	var flag atomic.Int64
	done := make(chan struct{})
	go func() {
		for i := 0; i < m; i++ {
			for flag.Load() != int64(2*i+1) {
			}
			flag.Store(int64(2*i + 2))
		}
		close(done)
	}()
	start := time.Now()
	for i := 0; i < m; i++ {
		flag.Store(int64(2*i + 1))
		for flag.Load() != int64(2*i+2) {
		}
	}
	d := time.Since(start)
	<-done
	return d / time.Duration(m)
}

// Gosched 往返：自旋里每次让出，调度器帮忙换人（1 个 P 即可）
func benchGoschedRound(m int) time.Duration {
	var flag atomic.Int64
	done := make(chan struct{})
	go func() {
		for i := 0; i < m; i++ {
			for flag.Load() != int64(2*i+1) {
				runtime.Gosched()
			}
			flag.Store(int64(2*i + 2))
		}
		close(done)
	}()
	start := time.Now()
	for i := 0; i < m; i++ {
		flag.Store(int64(2*i + 1))
		for flag.Load() != int64(2*i+2) {
			runtime.Gosched()
		}
	}
	d := time.Since(start)
	<-done
	return d / time.Duration(m)
}

// channel 往返：unbuffered channel 打乒乓球，每次交接都是真 park/唤醒
func benchChanRound(m int) time.Duration {
	ch := make(chan struct{})
	ack := make(chan struct{})
	done := make(chan struct{})
	go func() {
		for i := 0; i < m; i++ {
			<-ch
			ack <- struct{}{}
		}
		close(done)
	}()
	start := time.Now()
	for i := 0; i < m; i++ {
		ch <- struct{}{}
		<-ack
	}
	d := time.Since(start)
	<-done
	return d / time.Duration(m)
}

func runLadder(m int) {
	if runtime.GOMAXPROCS(0) >= 2 {
		fmt.Printf("原子标志往返（纯自旋，无调度）      : %6d ns/次\n", benchAtomicRound(m))
	} else {
		fmt.Println("原子标志往返（纯自旋，无调度）      : 跳过（需要 GOMAXPROCS≥2，否则就是实验 1）")
	}
	rt := benchGoschedRound(m)
	fmt.Printf("Gosched 往返（让出换人，1 个 P）    : %6d ns/次\n", rt)
	rt = benchChanRound(m)
	fmt.Printf("channel 往返（2 次 park/唤醒交接）  : %6d ns/次  ≈ %d ns/次 park/唤醒\n", rt, rt/2)
}

// ---------- main ----------

func main() {
	exp := flag.String("exp", "stall", "stall | ladder")
	procs := flag.Int("procs", 0, "GOMAXPROCS（0 = 不改）")
	n := flag.Int("n", 200000, "stall: 元素个数")
	capacity := flag.Int("cap", 1024, "stall: 队列容量（2 的幂）")
	sleepFor := flag.Duration("sleep", 100*time.Microsecond, "stall: Sleep 策略的睡眠时长")
	rounds := flag.Int("rounds", 3, "stall: 每种策略跑几轮")
	only := flag.Int("wait", -1, "stall: 只跑指定策略 0自旋/1Gosched/2Sleep（-1=全部）")
	m := flag.Int("m", 200000, "ladder: 往返次数")
	flag.Parse()

	if *procs > 0 {
		runtime.GOMAXPROCS(*procs)
	}
	fmt.Printf("Go %s | GOMAXPROCS=%d\n", runtime.Version(), runtime.GOMAXPROCS(0))

	switch *exp {
	case "stall":
		runStall(*n, *capacity, *sleepFor, *rounds, *only)
	case "ladder":
		runLadder(*m)
	default:
		fmt.Println("unknown -exp:", *exp)
	}
}
