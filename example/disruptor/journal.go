package main

import (
	"bufio"
	"encoding/binary"
	"hash/crc32"
	"os"
)

// Journal：顺序追加 + 每记录自校验（§4.1）。"磁盘是新的磁带"——
// 随机 I/O 是机械盘死穴，顺序追加快得离谱，所以 journal 只做 append
//
// 记录格式：len(4B) | crc32(4B) | payload —— 自描述、可校验、崩溃可截断
type Journal struct {
	f       *os.File
	w       *bufio.Writer
	written int64 // 已写字节数：快照要记"journal 重放到哪"，文章骨架之外补的账
}

func OpenJournal(path string) (*Journal, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	return &Journal{f: f, w: bufio.NewWriter(f)}, nil
}

func (j *Journal) Append(payload []byte) {
	var head [8]byte
	binary.LittleEndian.PutUint32(head[0:4], uint32(len(payload)))
	binary.LittleEndian.PutUint32(head[4:8], crc32.ChecksumIEEE(payload))
	j.w.Write(head[:])
	j.w.Write(payload)
	j.written += 8 + int64(len(payload))
}

// Written：快照点对应的 journal 偏移（§4.2 Snapshot 的第二个参数从这来）
func (j *Journal) Written() int64 { return j.written }

// Sync：只在 endOfBatch 里调——一批一次 fsync，丢失窗口 ≤ 一批，I/O 摊薄到批粒度（§3.2）
func (j *Journal) Sync() error {
	if err := j.w.Flush(); err != nil {
		return err
	}
	return j.f.Sync()
}
