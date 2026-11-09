package main

import (
	"encoding/binary"
	"fmt"
)

// Request / OutputEvent：业务对象（§1 Event 的载荷）。
// 生产者只带 Raw 进来，unmarshaler 把 Raw 反序列化写进同一槽位的 Req——每个字段只有一个写者
type Request struct {
	ID     int64
	Amount int64
}

type OutputEvent struct {
	Topic string
	Text  string
}

// State：BLP 的全内存状态（§4.1）。演示版两个计数器；
// 真实 LMAX 里是订单簿、账户、持仓——2GB 量级，但结构同款：Apply 是唯一入口
type State struct {
	Applied  int64
	Turnover int64
}

// Apply 必须是确定性的：同一事件序列 → 同一状态转移 + 同一输出。
// 这是整个架构的地基：输出不落盘，丢了重算就有；恢复 = 重放同一份 journal
func (s *State) Apply(req Request) []OutputEvent {
	s.Applied++
	s.Turnover += req.Amount
	return []OutputEvent{{Topic: "trades", Text: fmt.Sprintf("order-%d done", req.ID)}}
}

// marshal：快照用的固定 16B 布局（§4.2）。生产版上 gob / protobuf
func (s *State) marshal() []byte {
	b := make([]byte, 16)
	binary.LittleEndian.PutUint64(b[0:8], uint64(s.Applied))
	binary.LittleEndian.PutUint64(b[8:16], uint64(s.Turnover))
	return b
}

func unmarshalState(b []byte) *State {
	return &State{
		Applied:  int64(binary.LittleEndian.Uint64(b[0:8])),
		Turnover: int64(binary.LittleEndian.Uint64(b[8:16])),
	}
}

// marshalRequest / unmarshalRequest：journal 记录 payload ↔ Request。
// 与输入环的 Raw 同一格式（BigEndian uint64 × 2），journal 字节流因此能直接喂给重放
func marshalRequest(r Request) []byte {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[0:8], uint64(r.ID))
	binary.BigEndian.PutUint64(b[8:16], uint64(r.Amount))
	return b
}

func unmarshalRequest(payload []byte) Request {
	return Request{
		ID:     int64(binary.BigEndian.Uint64(payload[0:8])),
		Amount: int64(binary.BigEndian.Uint64(payload[8:16])),
	}
}
