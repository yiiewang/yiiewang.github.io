package main

// Topic：输出侧每主题一个独立 Disruptor，BLP 是它唯一的生产者（§6）。
// 下游一条依赖链：marshal（序列化）→ send（网络发送）
type Topic struct {
	d    *Disruptor
	sent *BatchProcessor // 链尾消费者：等它的进度就是等"回执出门"
}

// NewTopic：装配输出环。等待策略用 Yielding——输出侧是后台管线，让步省核（§3.1 档位）
func NewTopic(sizePow2 int, send func(OutputEvent)) *Topic {
	d := NewDisruptor(sizePow2, NewSingleProducer(int64(sizePow2)))
	marsh := d.HandleWith(Yielding{}, func(ev *Event, seq int64, eob bool) {
		ev.Out.Text = "topic:" + ev.Out.Text // 序列化占位：真实版是 protobuf / JSON 编码
	})
	sent := d.After([]*BatchProcessor{marsh}, Yielding{}, func(ev *Event, seq int64, eob bool) {
		send(ev.Out)
	})
	return &Topic{d: d, sent: sent}
}

// Publish：BLP 侧两步——先写槽位 Out 字段，后发布（§6 output.go 同款）。
// 只写 Out 不碰 Raw/Req：字段级唯一写者原则一路贯彻到输出侧
func (t *Topic) Publish(out OutputEvent) {
	seq := t.d.seqr.Next()
	t.d.rb.slot(seq).Out = out
	t.d.seqr.Publish(seq)
}
