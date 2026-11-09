package main

type Sequencer interface {
	Cursor() *Sequence                         // 发布边界，消费者看这里
	AddGating(s ...*Sequence)                  // 登记消费者：满环检查用
	Next() int64                               // 认领槽位（满则自旋等待）
	Publish(seq int64)                         // 数据就位后发布
	HighestPublished(lower, upper int64) int64 // 缺口检查，单生产者直接透传
}
