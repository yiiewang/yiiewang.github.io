package main

type Event struct {
	Raw []byte      // 网络原始字节：journaler / replicator / unmarshaler 读
	Req Request     // 业务对象：unmarshaler 写、BLP 读
	Out OutputEvent // 输出事件：BLP 写、输出侧 marshal 读
}

type RingBuffer struct {
	slots []Event
	mask  int64
}

func NewRingBuffer(sizePow2 int) *RingBuffer {
	return &RingBuffer{slots: make([]Event, sizePow2), mask: int64(sizePow2 - 1)}
}

func (rb *RingBuffer) slot(seq int64) *Event { return &rb.slots[seq&rb.mask] }
