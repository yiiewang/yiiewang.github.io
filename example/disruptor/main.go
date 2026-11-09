package main

// §6 总装的缩小版：单生产者输入环 → journal/replicate/unmarshal 并行组 → BLP 汇合。
// 收尾带两个小 demo：§6 输出侧（Topic 环）与 §4 事件溯源闭环（journal/快照/恢复）。
//
// 跑起来重点看五件事：
//   1. 依赖图：BLP 的进度永远追不过 min(journal, replica, unmarshal)——replica 刻意最慢，它说了算
//   2. 批处理：fsync 次数 << 事件数（endOfBatch 收口，I/O 摊到批粒度）
//   3. 背压：1024 槽的小环 + 最慢的 replica，生产者必然被挡进 Next() 的满环自旋
//   4. 输出侧两跳：Topic 环上 marshal → send，BLP 是输出环唯一生产者
//   5. 事件溯源：journal → 快照 → 尾部半条截断 → Recover 确定性追平
//
// 热路径里 journal / replicate 只做计数/耗时代价模拟；
// 真实的 journal 格式、快照、恢复在同目录 journal.go / snapshot.go / recover.go（§4）。

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"time"
)

// burn：模拟 replicator 的"推备节点 + 等 ACK"开销，刻意做成最慢消费者
func burn(b []byte) {
	var h int64
	for r := 0; r < 32; r++ {
		for _, x := range b {
			h = h*31 + int64(x)
		}
	}
	_ = h
}

func main() {
	const (
		bufSize = 1 << 10 // 1024 槽演示；LMAX 输入环 2000 万（§6 参数表）
		events  = 100_000
	)

	// 接线必须先于第一次 Publish：AddGating 改的是生产者的满环检查名单（§3.3 硬约束）
	d := NewDisruptor(bufSize, NewSingleProducer(int64(bufSize)))

	var (
		journaled  int64 // journaler 累计"落盘"字节
		fsyncs     int64 // journaler 批末 fsync 次数：O(批数)，远小于 O(事件数)
		replicated int64
	)

	// 并行组 1：三个消费者同时消费同一批事件
	journal := d.HandleWith(BusySpin{}, func(ev *Event, seq int64, eob bool) {
		atomic.AddInt64(&journaled, int64(len(ev.Raw)))
		if eob {
			atomic.AddInt64(&fsyncs, 1) // 一批一次 fsync：丢失窗口 ≤ 一批
		}
	})
	replica := d.HandleWith(BusySpin{}, func(ev *Event, seq int64, eob bool) {
		burn(ev.Raw)
		atomic.AddInt64(&replicated, 1)
	})
	unmarsh := d.HandleWith(BusySpin{}, func(ev *Event, seq int64, eob bool) {
		ev.Req = unmarshalRequest(ev.Raw) // 写本槽 Req 字段：此字段唯一写者
	})

	// BLP：三个前置全部就位才消费。依赖图是正确性约束——
	// "回执"（这里用累计成交额代替）之前，事件必须已落盘、已复制、已反序列化
	var (
		applied  int64 // 单线程业务核心的状态：只有 BLP 碰，普通变量即可
		turnover int64
	)
	blp := d.After([]*BatchProcessor{journal, replica, unmarsh}, BusySpin{},
		func(ev *Event, seq int64, eob bool) {
			turnover += ev.Req.Amount
			applied++
		})

	// 输入侧唯一生产者：§0 第 1 站的网络读取 goroutine。
	// 每事件独立缓冲：演示版 Publish 按值拷贝 Event，槽里的 Raw 引用它；
	// 真实姿势是认领后直接写槽内字段，热路径零分配（§7 GC 纪律）
	start := time.Now()
	for i := 1; i <= events; i++ {
		raw := make([]byte, 16)
		binary.BigEndian.PutUint64(raw[0:8], uint64(i))         // 订单 ID
		binary.BigEndian.PutUint64(raw[8:16], uint64(i%1000+1)) // 金额
		d.Publish(&Event{Raw: raw})
	}

	// 排空：等 BLP 的 gating 追平生产者——min(journal, replica, unmarshal) 到齐
	for blp.seq.Get() < int64(events) {
		runtime.Gosched()
	}
	elapsed := time.Since(start)

	fmt.Printf("events=%d  elapsed=%v  throughput=%.0f ev/s\n",
		events, elapsed, float64(events)/elapsed.Seconds())
	fmt.Printf("journal: %d B, %d 次 fsync（批粒度：%.0f 条/次）\n",
		atomic.LoadInt64(&journaled), atomic.LoadInt64(&fsyncs),
		float64(events)/float64(atomic.LoadInt64(&fsyncs)))
	fmt.Printf("replica: %d 条已复制\n", atomic.LoadInt64(&replicated))
	fmt.Printf("BLP: applied=%d  turnover=%d（单线程累计，无锁）\n", applied, turnover)

	demoOutputTopic()
	demoEventSourcing()

	// 演示到此直接退出：真实 LMAX 常驻运行。注意 Run 的 stop 检查在 WaitFor 之外，
	// BusySpin 里的消费者收不到关停信号——优雅关停需要可中断的 WaitStrategy（08-2）
}

// demoOutputTopic：§6 输出侧。BLP 是输出环唯一生产者，下游一条链 marshal → send
func demoOutputTopic() {
	fmt.Println("\n-- §6 输出侧：Topic 环（marshal → send 两跳）")

	var sent []OutputEvent // send 在自己的 goroutine 里写，排空后主 goroutine 才读
	topic := NewTopic(1<<10, func(out OutputEvent) { sent = append(sent, out) })

	blp := &State{}
	const n = 5
	for i := 1; i <= n; i++ {
		for _, out := range blp.Apply(Request{ID: int64(i), Amount: int64(100 + i)}) {
			topic.Publish(out) // Publish 内部：Next 认领 → 写槽 Out → 发布
		}
	}
	for topic.sent.seq.Get() < n { // 排空：等链尾 send 的进度
		runtime.Gosched()
	}
	for _, out := range sent {
		fmt.Println("  ", out.Text) // send 侧收到的最终形态
	}
}

// demoEventSourcing：§4 闭环。journal 8 条 → 第 5 条后打快照 → 尾部补半条记录模拟崩溃
// → Recover = 加载快照 + 重放 6-8 + crc 截断，必须追平崩溃前状态
func demoEventSourcing() {
	fmt.Println("\n-- §4 事件溯源闭环：journal → 快照 → 崩溃截断 → Recover")

	dir, err := os.MkdirTemp("", "lmax-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	jp := filepath.Join(dir, "journal.bin")
	sp := filepath.Join(dir, "snap.bin")

	live := &State{}
	j, err := OpenJournal(jp)
	if err != nil {
		panic(err)
	}
	for i := 1; i <= 8; i++ {
		if i == 6 { // 第 5 条落账后打快照：恢复只需重放 6-8
			if err := live.Snapshot(sp, j.Written()); err != nil {
				panic(err)
			}
		}
		req := Request{ID: int64(i), Amount: int64(100 + i)}
		live.Apply(req)
		j.Append(marshalRequest(req))
	}
	if err := j.Sync(); err != nil {
		panic(err)
	}

	// 崩溃现场：journal 尾部只有半条记录——头声明 payload 9B，实际只写进 3B
	f, err := os.OpenFile(jp, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	f.Write([]byte{9, 0, 0, 0, 0xFF, 0xFF, 0xFF, 0xFF, 1, 2, 3})
	f.Close()

	recovered := Recover(sp, jp)
	fmt.Printf("  live      : applied=%d turnover=%d\n", live.Applied, live.Turnover)
	fmt.Printf("  recovered : applied=%d turnover=%d（快照含前 5 条，重放 6-8，半条被截断）\n",
		recovered.Applied, recovered.Turnover)
	if *recovered != *live {
		panic("recover mismatch")
	}
	fmt.Println("  OK：确定性重放追平崩溃前状态")
}
