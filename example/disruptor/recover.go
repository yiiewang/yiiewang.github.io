package main

import (
	"bufio"
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

// loadSnapshot：校验 magic/version，返回快照状态 + journal 重放起点。
// 生产级加固（A/B 双代轮换、数据 crc、坏档退上一代）见 §4.2——演示版单文件够用
func loadSnapshot(snapPath string) (*State, int64) {
	b, err := os.ReadFile(snapPath)
	if err != nil {
		panic(err)
	}
	if len(b) < 24 ||
		binary.LittleEndian.Uint32(b[0:4]) != snapMagic ||
		binary.LittleEndian.Uint32(b[4:8]) != snapVersion {
		panic("snapshot magic/version mismatch")
	}
	off := int64(binary.LittleEndian.Uint64(b[8:16]))
	n := int64(binary.LittleEndian.Uint64(b[16:24]))
	if n < 0 || 24+n > int64(len(b)) {
		panic("snapshot truncated")
	}
	return unmarshalState(b[24 : 24+n]), off
}

// Recover：恢复三步——最近合法快照 → 重放其后的 journal → 半条截断（§4.3）。
// 崩溃一致性不靠重放过程完美，靠"每条记录独立校验 + 逐条重放天然终止于校验失败点"
func Recover(snapPath, journalPath string) *State {
	st, off := loadSnapshot(snapPath)
	f, _ := os.Open(journalPath)
	f.Seek(off, io.SeekStart) // 只重放快照点之后

	r := bufio.NewReader(f)
	var head [8]byte
	for {
		if _, err := io.ReadFull(r, head[:]); err != nil {
			break // 尾部不足 8B = 崩溃现场，截断
		}
		n := binary.LittleEndian.Uint32(head[0:4])
		sum := binary.LittleEndian.Uint32(head[4:8])
		payload := make([]byte, n)
		if _, err := io.ReadFull(r, payload); err != nil ||
			crc32.ChecksumIEEE(payload) != sum {
			break // 半条或坏条：crc 失败处截断，之前的全部有效
		}
		st.Apply(unmarshalRequest(payload)) // 重放：状态生效，输出丢弃
	}
	return st
}
