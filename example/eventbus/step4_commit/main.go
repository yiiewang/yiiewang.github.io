// 第 4 步：用事件总线模拟 cross 的两阶段提交，
// 并把"应答汇聚后只能提交一次"这个并发问题单独演示出来。
//
// 总线直接沿用第 2 步（异步 + 排空式 Close），本步没有用 Hook。
// 唯一的改动：Event.Data 从 string 改成 interface{}，用来携带结构化数据
// （cross 里的 event.Event 也是接口，处理器自己做类型断言）。
//
// 流程（与 cross/ARCHITECTURE.md 第 6 节一致）：
//
//	跨链事件发生 ──► 为每个子操作发 转发请求
//	转发请求     ──► 接收器转成 子链执行
//	子链执行     ──► 适配器调用子链合约（阶段一），发 子链应答
//	子链应答     ──► 接收器记录证明，汇聚：全部子事件都有应答 ──► 发 主链提交
//	主链提交     ──► 适配器调用主链 Commit 合约（阶段二）
//
// 与 cross 的对应关系：
//
//	AdapterHandler.onOccur      <-> adapter_handler.go  handleListenData / handleSubEvent
//	ReceiverHandler(Request)    <-> receiver.go         handleRequest
//	AdapterHandler.onExecute    <-> adapter_handler.go  crossEventExecute
//	ReceiverHandler.handleResp  <-> receiver.go         handleResponse（含 commitGuard）
//	AdapterHandler.onCommit     <-> adapter_handler.go  crossEventCommit
//
// 相对 cross 的简化：省略证明器（prover）；汇聚时当前应答也从存储读取，
// 而 cross 对"当前这条应答"使用内存中的证明上下文、其余从库里读。
//
// 并发从哪来：每个"子链应答"事件被总线放进各自独立的 goroutine，
// 多个子链几乎同时返回时，多个 handleResponse 会同时运行。
//
// 运行：go run -race ./step4_commit
//
// 观察重点：
//  1. 场景一（无防护）：三个应答都看到"已收齐 3/3"，于是主链 Commit 被调用 3 次
//  2. 场景二（原子占位）：同样三个应答都看到 3/3，但只有一个抢到占位，Commit 只调用 1 次
//  3. 链B 执行失败，它的失败应答同样参与汇聚（"完成"不等于"成功"）
//  4. go run -race 不会报警：每次访问都加了锁，这里的缺陷是逻辑竞态，
//     不是数据竞态。-race 抓不到它，只能靠推演并发交错来发现
package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// 模拟存储访问耗时。真实系统里读写库同样需要时间，这个时间窗口就是竞态的来源；
// 这里把它放大到毫秒级，让问题稳定复现、肉眼可见。
const (
	writeLatency = 5 * time.Millisecond
	readLatency  = 20 * time.Millisecond
	execBase     = 50 * time.Millisecond // 各子链执行耗时几乎相同 => 应答几乎同时到达
)

// 每个场景重新计时，所以用原子变量保存起点。
var startNs atomic.Int64

func resetClock() { startNs.Store(time.Now().UnixNano()) }

func logf(format string, args ...interface{}) {
	ms := (time.Now().UnixNano() - startNs.Load()) / int64(time.Millisecond)
	fmt.Printf("[%4dms] %s\n", ms, fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------- 总线（第 2 步，Data 改为 interface{}）

type EventType string

type Event struct {
	Type EventType
	Data interface{}
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

func (b *Bus) dispatchLoop() {
	for {
		select {
		case q := <-b.queue:
			for _, h := range q.hs {
				go func(h Handler) {
					defer b.wg.Done()
					if _, err := h.Handle(q.e); err != nil {
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

// ---------------------------------------------------------------- 事件与载荷

const (
	EvOccur    EventType = "跨链事件发生" // cross: CrossEventOccur
	EvRequest  EventType = "转发请求"   // cross: TransferRequestReceiveOpEvent
	EvExecute  EventType = "子链执行"   // cross: CrossEventExecute
	EvResponse EventType = "子链应答"   // cross: TransferResponseReceiveOpEvent
	EvCommit   EventType = "主链提交"   // cross: CrossEventCommit
)

// SubOp 一个子链操作。WillFail 仅用于演示：让该子链的合约执行失败。
type SubOp struct {
	Chain    string
	WillFail bool
}

type CrossEvent struct {
	CrossID string
	Subs    []SubOp
}

// SubTask 转发请求与子链执行共用的载荷。
type SubTask struct {
	CrossID string
	ExecIdx int
	SubOp
}

type Response struct {
	CrossID string
	ExecIdx int
	Chain   string
	OK      bool
}

type CommitReq struct {
	CrossID string
	Summary string
}

// payload 把事件载荷断言为期望类型，类型不符时返回错误而不是 panic。
func payload[T any](e *Event) (T, error) {
	v, ok := e.Data.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("event %q: unexpected payload %T", e.Type, e.Data)
	}
	return v, nil
}

// ---------------------------------------------------------------- 模拟的存储与主链

// SubEvent 对应数据库里的子事件记录。
type SubEvent struct {
	Chain  string
	Proved bool // 已记录应答（cross: StxProveStatus > Unknown）
	OK     bool // 子链合约是否执行成功
}

// Store 内存模拟的数据库。所有访问都加锁，所以没有数据竞态；
// 但"先写、再读全部、再判断"是三个独立步骤，中间可被别的协程插入。
type Store struct {
	mu     sync.Mutex
	subs   map[string][]SubEvent
	status map[string]string
}

func NewStore() *Store {
	return &Store{subs: map[string][]SubEvent{}, status: map[string]string{}}
}

func (s *Store) Create(crossID string, ops []SubOp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	subs := make([]SubEvent, len(ops))
	for i, op := range ops {
		subs[i] = SubEvent{Chain: op.Chain}
	}
	s.subs[crossID] = subs
	s.status[crossID] = "Init"
}

func (s *Store) MarkProved(crossID string, idx int, ok bool) {
	time.Sleep(writeLatency) // 写入落盘需要时间
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subs[crossID][idx].Proved = true
	s.subs[crossID][idx].OK = ok
}

// LoadSubs 读取全部子事件。读取耗时之后才取快照，
// 所以读到的是"耗时结束那一刻"的状态，期间别人的写入都可见。
func (s *Store) LoadSubs(crossID string) []SubEvent {
	time.Sleep(readLatency)
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]SubEvent(nil), s.subs[crossID]...)
}

func (s *Store) SetStatus(crossID, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status[crossID] = status
}

func (s *Store) Status(crossID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status[crossID]
}

// MainChain 模拟主链上的 Commit 合约，只记录被调用了几次。
// 真实的 Commit 合约通常不是幂等的：重复调用可能重复结算，或被合约拒绝并报错。
type MainChain struct {
	mu    sync.Mutex
	calls map[string]int
}

func NewMainChain() *MainChain { return &MainChain{calls: map[string]int{}} }

func (c *MainChain) InvokeCommit(crossID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls[crossID]++
	return c.calls[crossID]
}

func (c *MainChain) Calls(crossID string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls[crossID]
}

// ---------------------------------------------------------------- 处理器：适配器

// AdapterHandler 对应 cross 的 ChainAdapterHandler：负责与链交互。
type AdapterHandler struct {
	bus   *Bus
	store *Store
	chain *MainChain
}

func (h *AdapterHandler) Name() string { return "AdapterHandler" }
func (h *AdapterHandler) SupportEventTypes() []EventType {
	return []EventType{EvOccur, EvExecute, EvCommit}
}

func (h *AdapterHandler) Handle(e *Event) (interface{}, error) {
	switch e.Type {
	case EvOccur:
		return nil, h.onOccur(e)
	case EvExecute:
		return nil, h.onExecute(e)
	case EvCommit:
		return nil, h.onCommit(e)
	}
	return nil, fmt.Errorf("unsupported event %q", e.Type)
}

// onOccur：落库，状态改为处理中，为每个子操作发一个转发请求。
func (h *AdapterHandler) onOccur(e *Event) error {
	ce, err := payload[*CrossEvent](e)
	if err != nil {
		return err
	}
	h.store.Create(ce.CrossID, ce.Subs)
	h.store.SetStatus(ce.CrossID, "Processing")
	logf("[适配器] 跨链事件 %s 发生：落库 %d 个子操作，状态 Init -> Processing", ce.CrossID, len(ce.Subs))
	for i, op := range ce.Subs {
		task := &SubTask{CrossID: ce.CrossID, ExecIdx: i, SubOp: op}
		if _, err := h.bus.Pub(&Event{Type: EvRequest, Data: task}, AsyncSubPubMode); err != nil {
			return err
		}
	}
	return nil
}

// onExecute：阶段一，在子链上执行合约，然后无论成败都回一个应答。
func (h *AdapterHandler) onExecute(e *Event) error {
	t, err := payload[*SubTask](e)
	if err != nil {
		return err
	}
	// 各子链耗时只差 1ms 的量级，应答会几乎同时到达。
	time.Sleep(execBase + time.Duration(t.ExecIdx)*time.Millisecond)
	ok := !t.WillFail
	if ok {
		logf("[适配器] 阶段一：%s 子合约执行成功 (#%d)", t.Chain, t.ExecIdx)
	} else {
		logf("[适配器] 阶段一：%s 子合约执行失败 (#%d)，仍然回应答", t.Chain, t.ExecIdx)
	}
	resp := &Response{CrossID: t.CrossID, ExecIdx: t.ExecIdx, Chain: t.Chain, OK: ok}
	_, err = h.bus.Pub(&Event{Type: EvResponse, Data: resp}, AsyncSubPubMode)
	return err
}

// onCommit：阶段二，调用主链 Commit 合约。
func (h *AdapterHandler) onCommit(e *Event) error {
	c, err := payload[*CommitReq](e)
	if err != nil {
		return err
	}
	n := h.chain.InvokeCommit(c.CrossID)
	if n == 1 {
		logf("[适配器] 阶段二：调用主链 Commit 合约（第 %d 次），%s", n, c.Summary)
	} else {
		logf("[适配器] 阶段二：调用主链 Commit 合约（第 %d 次），重复提交！", n)
	}
	h.store.SetStatus(c.CrossID, "Complete")
	return nil
}

// ---------------------------------------------------------------- 处理器：接收器

// Mode 汇聚后是否做"只提交一次"的防护。
type Mode int

const (
	NoGuard   Mode = iota // 无防护：存在竞态
	WithGuard             // 原子占位：与 cross 的 commitGuard 一致
)

func (m Mode) String() string {
	if m == NoGuard {
		return "无防护"
	}
	return "原子占位"
}

// ReceiverHandler 对应 cross 的 ReceiveHandler + Receiver。
type ReceiverHandler struct {
	bus   *Bus
	store *Store
	mode  Mode
	// commitGuard 记录"已发布过 Commit 的 crossID"，对应 cross receiver.go 的同名字段。
	commitGuard sync.Map
}

func (h *ReceiverHandler) Name() string { return "ReceiverHandler" }
func (h *ReceiverHandler) SupportEventTypes() []EventType {
	return []EventType{EvRequest, EvResponse}
}

func (h *ReceiverHandler) Handle(e *Event) (interface{}, error) {
	switch e.Type {
	case EvRequest:
		return nil, h.handleRequest(e)
	case EvResponse:
		return nil, h.handleResponse(e)
	}
	return nil, fmt.Errorf("unsupported event %q", e.Type)
}

// handleRequest：（省略证明）直接转成子链执行事件。
func (h *ReceiverHandler) handleRequest(e *Event) error {
	t, err := payload[*SubTask](e)
	if err != nil {
		return err
	}
	logf("[接收器] 收到转发请求 #%d -> %s，转为执行事件", t.ExecIdx, t.Chain)
	_, err = h.bus.Pub(&Event{Type: EvExecute, Data: t}, AsyncSubPubMode)
	return err
}

// handleResponse：汇聚的核心。三步：写自己 -> 读全部 -> 判断是否收齐并提交。
func (h *ReceiverHandler) handleResponse(e *Event) error {
	resp, err := payload[*Response](e)
	if err != nil {
		return err
	}

	// 第 1 步：记录自己这条应答。
	h.store.MarkProved(resp.CrossID, resp.ExecIdx, resp.OK)

	// 第 2 步：读取全部子事件，统计已有应答的个数。
	subs := h.store.LoadSubs(resp.CrossID)
	proved, okCount := 0, 0
	for _, s := range subs {
		if s.Proved {
			proved++
			if s.OK {
				okCount++
			}
		}
	}
	logf("[接收器] 应答 #%d(%s) 处理中：已收齐 %d/%d", resp.ExecIdx, resp.Chain, proved, len(subs))
	if proved != len(subs) {
		return nil // 还没收齐，等别的应答来触发
	}

	// 第 3 步：收齐了，发布 Commit。
	//
	// 问题在这里：第 1、2 步与第 3 步之间没有任何互斥。
	// 多个应答几乎同时到达时，每个协程都在自己的"读全部"里看到了 3/3，
	// 于是每个协程都认为"我是最后一个"，各发一次 Commit。
	//
	// 修复：把"是否已经提交过"的检查和标记合成一个原子操作。
	// 不能写成"先 Load 判断、再 Store 标记"——那仍是先检查后行动，
	// 两个协程可能都 Load 到"不存在"，再各自 Store。
	// LoadOrStore 保证只有一个调用方得到 loaded=false。
	if h.mode == WithGuard {
		if _, loaded := h.commitGuard.LoadOrStore(resp.CrossID, struct{}{}); loaded {
			logf("[接收器] 应答 #%d(%s)：该跨链事件的 Commit 已被别人发布，跳过", resp.ExecIdx, resp.Chain)
			return nil
		}
	}
	summary := fmt.Sprintf("汇总：成功 %d / 失败 %d", okCount, len(subs)-okCount)
	logf("[接收器] 应答 #%d(%s)：发布 Commit 事件（%s）", resp.ExecIdx, resp.Chain, summary)
	_, err = h.bus.Pub(&Event{Type: EvCommit, Data: &CommitReq{CrossID: resp.CrossID, Summary: summary}}, AsyncSubPubMode)
	return err
}

// ---------------------------------------------------------------- 运行

const crossID = "X1"

// run 跑一个完整的跨链事务，返回主链 Commit 合约被调用的次数。
func run(mode Mode) int {
	resetClock()
	bus := NewBus()
	store, chain := NewStore(), NewMainChain()
	for _, h := range []Handler{
		&AdapterHandler{bus: bus, store: store, chain: chain},
		&ReceiverHandler{bus: bus, store: store, mode: mode},
	} {
		if err := bus.RegisterHandler(h); err != nil {
			panic(err)
		}
	}

	ev := &CrossEvent{CrossID: crossID, Subs: []SubOp{
		{Chain: "链A"},
		{Chain: "链B", WillFail: true},
		{Chain: "链C"},
	}}
	if _, err := bus.Pub(&Event{Type: EvOccur, Data: ev}, AsyncSubPubMode); err != nil {
		panic(err)
	}
	bus.Close() // 等待整条事件链（含链式产生的事件）处理完

	n := chain.Calls(crossID)
	logf("结束：跨链事件状态=%s，主链 Commit 调用次数=%d", store.Status(crossID), n)
	return n
}

func main() {
	fmt.Println("===== 场景一：无防护 =====")
	bad := run(NoGuard)

	fmt.Println()
	fmt.Println("===== 场景二：原子占位（LoadOrStore）=====")
	good := run(WithGuard)

	fmt.Println()
	fmt.Println("===== 结果对比（子事件共 3 个，期望 Commit 只调用 1 次）=====")
	fmt.Printf("  %s：Commit 调用 %d 次\n", NoGuard, bad)
	fmt.Printf("  %s：Commit 调用 %d 次\n", WithGuard, good)

	if bad <= 1 {
		fmt.Println("提示：本次没有复现竞态（极少见）。再运行一次，或调大 readLatency。")
	}
	if good != 1 {
		fmt.Println("错误：加了防护仍不是恰好 1 次，请检查实现。")
		os.Exit(1)
	}
}
