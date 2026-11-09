// 第 2 步：加入异步 —— 发布立即返回，处理在后台进行。
//
// 相对第 1 步的变化（可用 diff step1_sync/main.go step2_async/main.go 对照）：
//  1. 新增 PubMode：SyncPubMode / AsyncSubPubMode
//  2. 新增队列 queue 和分发协程 dispatchLoop
//  3. 新增 WaitGroup 与 Close：等待在途事件处理完
//  4. Pub 返回 (结果, 错误)；异步模式的返回值只代表"已入队"
//
// 与 cross 的对应关系（cross/module/bus/hook_bus.go）：
//
//	queue + dispatchLoop  <-> channels + init()
//	SyncPubMode           <-> SyncPubMode：必须恰好一个处理器，当场返回结果
//	AsyncSubPubMode       <-> AsyncSubPubMode：每个处理器独立 goroutine，不保证顺序
//
// 运行：go run ./step2_async
//
// 观察重点：
//   - 3 个"下单"的 Pub 在 0ms 就全部返回了，但处理要几百毫秒
//   - 完成顺序是 B、C、A，不是发布顺序 —— 异步不保证顺序
//   - Close 会等到整条链路（含链式产生的新事件）处理完
//
// 注意：cross 的 Close 只等待 AsyncSub 处理协程，不排空队列；这里的 Close
// 通过"发布时先计数"做到了排空，便于理解这个差异。
package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var start = time.Now()

func logf(format string, args ...interface{}) {
	fmt.Printf("[%4dms] %s\n", time.Since(start).Milliseconds(), fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------- 总线

type EventType string

type Event struct {
	Type EventType
	Data string
}

type Handler interface {
	Name() string
	SupportEventTypes() []EventType
	Handle(e *Event) (interface{}, error)
}

type PubMode int

const (
	// SyncPubMode 同步：在发布者的调用里直接执行，返回处理器的结果。
	SyncPubMode PubMode = iota
	// AsyncSubPubMode 异步：入队后立即返回，每个处理器各起一个 goroutine。
	AsyncSubPubMode
)

var ErrBusClosed = errors.New("event bus has closed")

// queued 是放进队列的一条记录：事件 + 发布时确定的处理器列表。
type queued struct {
	e  *Event
	hs []Handler
}

type Bus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler

	queue     chan queued
	closeC    chan struct{}
	closeOnce sync.Once
	// wg 记录"已发布但尚未处理完"的 (事件, 处理器) 数量。
	// 必须在发布时（入队之前）就 Add，Close 才能保证等到它们。
	wg sync.WaitGroup
}

func NewBus() *Bus {
	b := &Bus{
		handlers: make(map[EventType][]Handler),
		queue:    make(chan queued, 64),
		closeC:   make(chan struct{}),
	}
	go b.dispatchLoop()
	return b
}

func (b *Bus) RegisterHandler(h Handler) error {
	types := h.SupportEventTypes()
	if len(types) == 0 {
		return errors.New("handler supports no event type")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, t := range types {
		b.handlers[t] = append(b.handlers[t], h)
	}
	return nil
}

func (b *Bus) handlersOf(t EventType) []Handler {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return append([]Handler(nil), b.handlers[t]...)
}

// Pub 发布事件。
//   - SyncPubMode：返回处理器的真实结果和错误
//   - AsyncSubPubMode：只返回"是否成功入队"，处理器的结果和错误发布者拿不到
func (b *Bus) Pub(e *Event, mode PubMode) (interface{}, error) {
	hs := b.handlersOf(e.Type)

	if mode == SyncPubMode {
		// 与 cross 一致：同步模式要求恰好一个处理器，否则"返回谁的结果"没有意义。
		if len(hs) != 1 {
			return nil, fmt.Errorf("sync mode needs exactly 1 handler for %q, got %d", e.Type, len(hs))
		}
		return hs[0].Handle(e)
	}

	if len(hs) == 0 {
		return nil, fmt.Errorf("no handler for event %q", e.Type)
	}
	select {
	case <-b.closeC:
		return nil, ErrBusClosed
	default:
	}
	b.wg.Add(len(hs))
	select {
	case b.queue <- queued{e: e, hs: hs}:
		return nil, nil
	case <-b.closeC:
		b.wg.Add(-len(hs))
		return nil, ErrBusClosed
	}
}

// dispatchLoop 独立的分发协程：从队列取事件，为每个处理器启动 goroutine。
// 自己只负责"派活"，所以队列总能被及时消化，发布者不会因处理慢而被卡住。
func (b *Bus) dispatchLoop() {
	for {
		select {
		case q := <-b.queue:
			logf("[总线] 取出事件 %q，分发给 %d 个处理器", q.e.Type, len(q.hs))
			for _, h := range q.hs {
				go func(h Handler) {
					defer b.wg.Done()
					if _, err := h.Handle(q.e); err != nil {
						// 异步模式下错误无人可返回，只能记日志（真实系统还应落库/重试）。
						logf("[总线] %s 处理 %q 出错: %v", h.Name(), q.e.Type, err)
					}
				}(h)
			}
		case <-b.closeC:
			return
		}
	}
}

// Close 先等待所有在途事件处理完，再停止分发协程。
// 处理器在处理中发布的新事件，会在它自己 Done 之前完成计数，因此也会被等到。
// 约束：Close 期间不应有外部协程继续发布。
func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		b.wg.Wait()
		close(b.closeC)
	})
}

// ---------------------------------------------------------------- 业务处理器

// StockHandler：收到"下单"，扣库存（耗时由商品决定），再异步宣布"库存已扣"。
type StockHandler struct{ bus *Bus }

var costMs = map[string]time.Duration{"A": 300, "B": 100, "C": 200}

func (h *StockHandler) Name() string                   { return "StockHandler" }
func (h *StockHandler) SupportEventTypes() []EventType { return []EventType{"下单"} }
func (h *StockHandler) Handle(e *Event) (interface{}, error) {
	logf("[库存] 开始扣减: 商品%s", e.Data)
	time.Sleep(costMs[e.Data] * time.Millisecond)
	logf("[库存] 扣减完成: 商品%s", e.Data)
	_, err := h.bus.Pub(&Event{Type: "库存已扣", Data: e.Data}, AsyncSubPubMode)
	return nil, err
}

// ShipHandler：收到"库存已扣"，发货。
type ShipHandler struct{}

func (h *ShipHandler) Name() string                   { return "ShipHandler" }
func (h *ShipHandler) SupportEventTypes() []EventType { return []EventType{"库存已扣"} }
func (h *ShipHandler) Handle(e *Event) (interface{}, error) {
	logf("[发货] 商品%s 已发货", e.Data)
	return nil, nil
}

// QueryHandler：同步查询，调用方需要立刻拿到结果。
type QueryHandler struct{}

func (h *QueryHandler) Name() string                   { return "QueryHandler" }
func (h *QueryHandler) SupportEventTypes() []EventType { return []EventType{"查询库存"} }
func (h *QueryHandler) Handle(e *Event) (interface{}, error) {
	return fmt.Sprintf("商品%s 库存=10", e.Data), nil
}

// ---------------------------------------------------------------- 运行

func main() {
	bus := NewBus()
	for _, h := range []Handler{&StockHandler{bus: bus}, &ShipHandler{}, &QueryHandler{}} {
		if err := bus.RegisterHandler(h); err != nil {
			panic(err)
		}
	}

	// 场景一：同步 —— 需要结果时使用，当场拿到返回值。
	res, err := bus.Pub(&Event{Type: "查询库存", Data: "A"}, SyncPubMode)
	logf("main: 同步查询结果 = %v, err = %v", res, err)

	// 场景二：异步 —— 三个下单事件连续发布，Pub 立即返回。
	for _, p := range []string{"A", "B", "C"} {
		if _, err := bus.Pub(&Event{Type: "下单", Data: p}, AsyncSubPubMode); err != nil {
			logf("main: 发布失败: %v", err)
		}
	}
	logf("main: 3 个 Pub 已全部返回（只代表入队，处理还没完成）")

	bus.Close()
	logf("main: Close 返回（所有在途事件，含链式产生的，已处理完）")

	// 场景三：关闭后再发布会被拒绝。
	_, err = bus.Pub(&Event{Type: "下单", Data: "A"}, AsyncSubPubMode)
	logf("main: 关闭后发布: %v", err)
}
