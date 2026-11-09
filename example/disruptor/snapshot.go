package main

import (
	"encoding/binary"
	"os"
)

const (
	snapMagic   = uint32(0x534E5031) // "SNP1"
	snapVersion = uint32(1)
)

// Snapshot：写临时文件 → fsync → rename 原子转正（§4.2，08-5 的三步舞）。
// rename 之前崩溃：旧快照永远完好；之后崩溃：新快照完整——不存在"半个快照"
//
// 快照内容 = 序列化状态 + journal 已重放到的偏移量：
// 恢复 = 加载快照 + 从该偏移重放，重放窗口从"全量日志"压回"一段日志"
func (s *State) Snapshot(snapPath string, journalOffset int64) error {
	tmp := snapPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()
	data := s.marshal()
	var head [24]byte
	binary.LittleEndian.PutUint32(head[0:4], snapMagic)
	binary.LittleEndian.PutUint32(head[4:8], snapVersion)
	binary.LittleEndian.PutUint64(head[8:16], uint64(journalOffset))
	binary.LittleEndian.PutUint64(head[16:24], uint64(len(data)))
	f.Write(head[:])
	f.Write(data)
	if err := f.Sync(); err != nil {
		return err
	}
	return os.Rename(tmp, snapPath) // 原子切换：崩溃前，旧快照永远完好
}
