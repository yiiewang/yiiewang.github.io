// 第 1 步：同步事件总线 —— 事件驱动的最小骨架。
//
// 与 cross 的对应关系：
//
//	Event      <-> event.OpEvent
//	Handler    <-> handler.EventHandler
//	Bus        <-> bus.HookEventBus（RegisterHandler + PubEvent 的同步模式）
//
// 运行：go run ./step1_sync
//
// 观察重点：Pub 是"嵌套调用"。订单处理器发布事件后，会等整条链路跑完才返回，
// 所以日志里"发货"出现在"短信通知"之前。
package main

import (
	"errors"
	"fmt"
	"time"
)

var start = time.Now()

func logf(format string, args ...interface{}) {
	fmt.Printf("[%4dms] %s\n", time.Since(start).Milliseconds(), fmt.Sprintf(format, args...))
}

// ---------------------------------------------------------------- 总线

type EventType string

// Event 是"发生了什么"的消息。Type 决定谁来处理，Data 是携带的内容。
type Event struct {
	Type EventType
	Data string
}

// Handler 处理器：声明自己关心哪些事件类型，并实现处理逻辑。
type Handler interface {
	Name() string
	SupportEventTypes() []EventType
	Handle(e *Event) (interface{}, error)
}

// Bus 本质就是一张 map[事件类型][]处理器 的登记表。
type Bus struct {
	handlers map[EventType][]Handler
}

func NewBus() *Bus {
	return &Bus{handlers: make(map[EventType][]Handler)}
}

// RegisterHandler 按处理器声明的类型登记。
func (b *Bus) RegisterHandler(h Handler) error {
	types := h.SupportEventTypes()
	if len(types) == 0 {
		return errors.New("handler supports no event type")
	}
	for _, t := range types {
		b.handlers[t] = append(b.handlers[t], h)
	}
	return nil
}

// Pub 发布事件：查表，依次调用所有登记的处理器，全部执行完才返回。
func (b *Bus) Pub(e *Event) error {
	hs := b.handlers[e.Type]
	if len(hs) == 0 {
		return fmt.Errorf("no handler for event %q", e.Type)
	}
	for _, h := range hs {
		if _, err := h.Handle(e); err != nil {
			return fmt.Errorf("handler %s: %w", h.Name(), err)
		}
	}
	return nil
}

// ---------------------------------------------------------------- 业务处理器
// 每个处理器只认识总线和事件，互相之间没有直接调用。

// OrderHandler：收到"下单"，创建订单，然后宣布"订单已创建"。
type OrderHandler struct{ bus *Bus }

func (h *OrderHandler) Name() string                   { return "OrderHandler" }
func (h *OrderHandler) SupportEventTypes() []EventType { return []EventType{"下单"} }
func (h *OrderHandler) Handle(e *Event) (interface{}, error) {
	logf("[订单] 创建订单: %s", e.Data)
	return nil, h.bus.Pub(&Event{Type: "订单已创建", Data: e.Data})
}

// StockHandler：收到"订单已创建"，扣库存，然后宣布"库存已扣"。
type StockHandler struct{ bus *Bus }

func (h *StockHandler) Name() string                   { return "StockHandler" }
func (h *StockHandler) SupportEventTypes() []EventType { return []EventType{"订单已创建"} }
func (h *StockHandler) Handle(e *Event) (interface{}, error) {
	logf("[库存] 扣减库存: %s", e.Data)
	return nil, h.bus.Pub(&Event{Type: "库存已扣", Data: e.Data})
}

// ShipHandler：收到"库存已扣"，发货。
type ShipHandler struct{}

func (h *ShipHandler) Name() string                   { return "ShipHandler" }
func (h *ShipHandler) SupportEventTypes() []EventType { return []EventType{"库存已扣"} }
func (h *ShipHandler) Handle(e *Event) (interface{}, error) {
	logf("[发货] 安排发货: %s", e.Data)
	return nil, nil
}

// NotifyHandler：也关心"订单已创建"，发短信。
// 这是事件驱动的核心好处：新增订阅者不需要修改 OrderHandler 的任何代码。
type NotifyHandler struct{}

func (h *NotifyHandler) Name() string                   { return "NotifyHandler" }
func (h *NotifyHandler) SupportEventTypes() []EventType { return []EventType{"订单已创建"} }
func (h *NotifyHandler) Handle(e *Event) (interface{}, error) {
	logf("[短信] 通知用户: %s", e.Data)
	return nil, nil
}

// PointsHandler：同样关心"订单已创建"，给用户发放积分。
// 新增这个处理器只做了两件事：实现 Handler 接口，并在 main 里注册；
// OrderHandler、StockHandler 等已有代码一行都没改。
type PointsHandler struct{}

func (h *PointsHandler) Name() string                   { return "PointsHandler" }
func (h *PointsHandler) SupportEventTypes() []EventType { return []EventType{"订单已创建"} }
func (h *PointsHandler) Handle(e *Event) (interface{}, error) {
	logf("[积分] 发放积分: %s", e.Data)
	return nil, nil
}

// ---------------------------------------------------------------- 运行

func main() {
	bus := NewBus()
	// 注册顺序决定同一事件下处理器的调用顺序：Stock -> Notify -> Points。
	for _, h := range []Handler{
		&OrderHandler{bus: bus},
		&StockHandler{bus: bus},
		&NotifyHandler{},
		&PointsHandler{},
		&ShipHandler{},
	} {
		if err := bus.RegisterHandler(h); err != nil {
			panic(err)
		}
	}

	logf("main: 发布 [下单]")
	if err := bus.Pub(&Event{Type: "下单", Data: "商品A"}); err != nil {
		logf("main: 出错: %v", err)
	}
	logf("main: Pub 返回（此时整条链路已全部执行完）")
}
