// 第 3 步：加入 Hook —— 在处理器执行前后插入"检查站"。
//
// 相对第 2 步的变化（可用 diff step2_async/main.go step3_hook/main.go 对照）：
//  1. 新增 Hook / HookManager / HookContext
//  2. Bus 新增 hooks 字段与 handleEvent：所有处理器调用都经过它
//  3. 业务处理器代码完全不变，Hook 从外部介入
//
// 与 cross 的对应关系：
//
//	Hook / HookContext / HookPoint / HookAction <-> cross/module/hook/hook.go
//	HookManager                                  <-> cross/hook_manager.go
//	Bus.handleEvent                              <-> cross/module/bus/hook_bus.go 的 handleEvent
//
// 运行：go run ./step3_hook
//
// 观察重点：
//  1. Hook 按 Priority 从小到大执行（黑名单1 -> 故障Hook5 -> 审计100）
//  2. Hook 返回错误：只记录，继续执行下一个 Hook 和处理器
//  3. Hook 返回 Reject：短路，后面的 Hook 和处理器都不再执行
//     所以被拒绝的事件，优先级更低的审计 Hook 看不到它
//  4. 后置 Hook 能拿到处理器的结果和错误（即使出错）
//  5. 同步调用时，调用方依然能拿到处理器的真实错误
package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var start = time.Now()

func logf(format string, args ...interface{}) {
	fmt.Printf("[%4dms] %s\n", time.Since(start).Milliseconds(), fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------- Hook 框架

type HookPoint int

const (
	BeforeHandle HookPoint = iota // 处理器执行前
	AfterHandle                   // 处理器执行后
)

func (p HookPoint) String() string {
	if p == BeforeHandle {
		return "前置"
	}
	return "后置"
}

type HookAction int

const (
	ContinueAction HookAction = iota // 继续流转
	RejectAction                     // 拒绝：不再执行后续 Hook 与处理器
)

// HookContext 传给 Hook 的上下文。Result/Err 仅在后置点有意义。
type HookContext struct {
	Event       *Event
	HandlerName string
	Point       HookPoint
	Result      interface{}
	Err         error
}

type Hook interface {
	Name() string
	// Priority 数值越小越先执行。
	Priority() int
	SupportEventTypes() []EventType
	SupportHookPoints() []HookPoint
	Execute(ctx *HookContext) (HookAction, error)
}

// HookManager 管理 Hook。cross 按 事件类型×触发点 建了二级索引加速查找，
// 这里为了好懂，用一个按优先级排序的切片，执行时过滤，语义相同。
type HookManager struct {
	mu    sync.RWMutex
	hooks []Hook
}

func NewHookManager() *HookManager { return &HookManager{} }

func (m *HookManager) Register(h Hook) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, old := range m.hooks {
		if old.Name() == h.Name() {
			return fmt.Errorf("hook %q already registered", h.Name())
		}
	}
	m.hooks = append(m.hooks, h)
	sort.SliceStable(m.hooks, func(i, j int) bool { return m.hooks[i].Priority() < m.hooks[j].Priority() })
	return nil
}

// Execute 依次执行匹配的 Hook，返回"是否继续事件流转"。
func (m *HookManager) Execute(ctx *HookContext) bool {
	m.mu.RLock()
	hooks := append([]Hook(nil), m.hooks...)
	m.mu.RUnlock()

	for _, h := range hooks {
		if !supportsType(h, ctx.Event.Type) || !supportsPoint(h, ctx.Point) {
			continue
		}
		action, err := h.Execute(ctx)
		if err != nil {
			// Hook 自己出错不能拖垮主流程：记录后继续下一个。
			logf("[Hook管理器] %s 执行出错(已忽略): %v", h.Name(), err)
			continue
		}
		if action == RejectAction {
			logf("[Hook管理器] %s 拒绝了事件 %q", h.Name(), ctx.Event.Type)
			return false
		}
	}
	return true
}

func supportsType(h Hook, t EventType) bool {
	for _, x := range h.SupportEventTypes() {
		if x == t {
			return true
		}
	}
	return false
}

func supportsPoint(h Hook, p HookPoint) bool {
	for _, x := range h.SupportHookPoints() {
		if x == p {
			return true
		}
	}
	return false
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
	SyncPubMode PubMode = iota
	AsyncSubPubMode
)

var ErrBusClosed = errors.New("event bus has closed")

type queued struct {
	e  *Event
	hs []Handler
}

type Bus struct {
	mu       sync.RWMutex
	handlers map[EventType][]Handler
	hooks    *HookManager // 为 nil 时等价于没有 Hook

	queue     chan queued
	closeC    chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
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

// SetHookManager 须在开始发布事件之前调用。
func (b *Bus) SetHookManager(m *HookManager) { b.hooks = m }

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

func (b *Bus) Pub(e *Event, mode PubMode) (interface{}, error) {
	hs := b.handlersOf(e.Type)

	if mode == SyncPubMode {
		if len(hs) != 1 {
			return nil, fmt.Errorf("sync mode needs exactly 1 handler for %q, got %d", e.Type, len(hs))
		}
		return b.handleEvent(e, hs[0]) // 与第 2 步的唯一区别：不再直接 Handle
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

// handleEvent 是所有处理器调用的唯一入口：前置 Hook -> 处理器 -> 后置 Hook。
func (b *Bus) handleEvent(e *Event, h Handler) (interface{}, error) {
	if b.hooks == nil {
		return h.Handle(e)
	}
	ctx := &HookContext{Event: e, HandlerName: h.Name(), Point: BeforeHandle}
	if !b.hooks.Execute(ctx) {
		// 被拒绝：处理器不会被调用，也不会触发后置 Hook。
		logf("[总线] 事件 %q 被前置 Hook 拒绝，%s 不会执行", e.Type, h.Name())
		return nil, nil
	}

	result, err := h.Handle(e)

	ctx.Point, ctx.Result, ctx.Err = AfterHandle, result, err
	b.hooks.Execute(ctx) // 后置 Hook 的结果不影响返回值
	return result, err
}

func (b *Bus) dispatchLoop() {
	for {
		select {
		case q := <-b.queue:
			for _, h := range q.hs {
				go func(h Handler) {
					defer b.wg.Done()
					if _, err := b.handleEvent(q.e, h); err != nil {
						logf("[总线] %s 处理 %q 出错: %v", h.Name(), q.e.Type, err)
					}
				}(h)
			}
		case <-b.closeC:
			return
		}
	}
}

func (b *Bus) Close() {
	b.closeOnce.Do(func() {
		b.wg.Wait()
		close(b.closeC)
	})
}

// ---------------------------------------------------------------- 业务处理器（与 Hook 无关）

// OrderHandler：处理"下单"。商品为"缺货商品"时返回错误，否则宣布"库存已扣"。
type OrderHandler struct{ bus *Bus }

func (h *OrderHandler) Name() string                   { return "OrderHandler" }
func (h *OrderHandler) SupportEventTypes() []EventType { return []EventType{"下单"} }
func (h *OrderHandler) Handle(e *Event) (interface{}, error) {
	if e.Data == "缺货商品" {
		return nil, errors.New("库存不足")
	}
	logf("[订单] 下单成功: %s", e.Data)
	if _, err := h.bus.Pub(&Event{Type: "库存已扣", Data: e.Data}, AsyncSubPubMode); err != nil {
		return nil, err
	}
	return "订单号-" + e.Data, nil
}

type ShipHandler struct{}

func (h *ShipHandler) Name() string                   { return "ShipHandler" }
func (h *ShipHandler) SupportEventTypes() []EventType { return []EventType{"库存已扣"} }
func (h *ShipHandler) Handle(e *Event) (interface{}, error) {
	logf("[发货] 已发货: %s", e.Data)
	return nil, nil
}

// ---------------------------------------------------------------- 三个示例 Hook

// BlacklistHook：前置拦截，违禁品直接拒绝。优先级最高（数值最小）。
type BlacklistHook struct{}

func (BlacklistHook) Name() string                   { return "黑名单Hook" }
func (BlacklistHook) Priority() int                  { return 1 }
func (BlacklistHook) SupportEventTypes() []EventType { return []EventType{"下单"} }
func (BlacklistHook) SupportHookPoints() []HookPoint { return []HookPoint{BeforeHandle} }
func (BlacklistHook) Execute(ctx *HookContext) (HookAction, error) {
	logf("[黑名单Hook] 检查商品: %s", ctx.Event.Data)
	if ctx.Event.Data == "违禁品" {
		return RejectAction, nil
	}
	return ContinueAction, nil
}

// FlakyHook：故意出错，演示"Hook 出错不影响主流程"。
type FlakyHook struct{}

func (FlakyHook) Name() string                   { return "故障Hook" }
func (FlakyHook) Priority() int                  { return 5 }
func (FlakyHook) SupportEventTypes() []EventType { return []EventType{"下单"} }
func (FlakyHook) SupportHookPoints() []HookPoint { return []HookPoint{BeforeHandle} }
func (FlakyHook) Execute(ctx *HookContext) (HookAction, error) {
	return ContinueAction, errors.New("我出故障了")
}

// AuditHook：前后置都记录，优先级最低（最后执行）。
type AuditHook struct{}

func (AuditHook) Name() string  { return "审计Hook" }
func (AuditHook) Priority() int { return 100 }
func (AuditHook) SupportEventTypes() []EventType {
	return []EventType{"下单", "库存已扣"}
}
func (AuditHook) SupportHookPoints() []HookPoint {
	return []HookPoint{BeforeHandle, AfterHandle}
}
func (AuditHook) Execute(ctx *HookContext) (HookAction, error) {
	if ctx.Point == BeforeHandle {
		logf("[审计Hook] %s 即将处理 %q(%s)", ctx.HandlerName, ctx.Event.Type, ctx.Event.Data)
	} else {
		logf("[审计Hook] %s 处理完 %q(%s): result=%v err=%v",
			ctx.HandlerName, ctx.Event.Type, ctx.Event.Data, ctx.Result, ctx.Err)
	}
	return ContinueAction, nil
}

// ---------------------------------------------------------------- 运行

func main() {
	bus := NewBus()

	hm := NewHookManager()
	for _, h := range []Hook{AuditHook{}, FlakyHook{}, BlacklistHook{}} { // 故意乱序注册
		if err := hm.Register(h); err != nil {
			panic(err)
		}
	}
	bus.SetHookManager(hm)

	for _, h := range []Handler{&OrderHandler{bus: bus}, &ShipHandler{}} {
		if err := bus.RegisterHandler(h); err != nil {
			panic(err)
		}
	}

	// 用同步模式发布"下单"，是为了让调用方也能看到处理器的结果/错误；
	// 其中"库存已扣"仍是异步链式事件。sleep 仅用于让日志顺序易读。
	for _, product := range []string{"正常商品", "违禁品", "缺货商品"} {
		fmt.Println()
		logf("main: ====== 下单 %s ======", product)
		res, err := bus.Pub(&Event{Type: "下单", Data: product}, SyncPubMode)
		logf("main: 调用方拿到 result=%v err=%v", res, err)
		time.Sleep(50 * time.Millisecond)
	}

	bus.Close()
	fmt.Println()
	logf("main: 全部结束")
}
