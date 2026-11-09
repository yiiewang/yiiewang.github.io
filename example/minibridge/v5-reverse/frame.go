package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// 帧类型。v2 的多路复用就靠这几种帧撑起一条流的完整生命周期；
// v5 在会话级上追加控制面三帧，撑起反向转发的注册与注销。
const (
	frameSYN    frameType = 1 // 请求开流，payload 是目标地址
	frameSYNACK frameType = 2 // 同意开流
	frameRST    frameType = 3 // 拒绝或立即终止
	frameDATA   frameType = 4 // 业务数据
	frameFIN    frameType = 5 // 半关闭：我发完了，但还能收
	framePING   frameType = 6 // 保活探测（v4 心跳启用）
	framePONG   frameType = 7 // 保活回应（v4 心跳启用）

	// 控制面三帧都是会话级（streamID=0），不查流表。
	// LISTEN/UNLISTEN 由客户端发给服务端，CTLRESULT 由服务端回应。
	// 每条都带请求 ID：请求方超时放弃后，迟到的旧回应不能冒充新答案。
	frameLISTEN    frameType = 8  // 请对端监听，payload "id|bind|target"
	frameUNLISTEN  frameType = 9  // 注销监听，payload "id|bind"
	frameCTLRESULT frameType = 10 // 请求结果，payload "id|ok" / "id|err: ..."
)

type frameType byte

func (t frameType) String() string {
	switch t {
	case frameSYN:
		return "SYN"
	case frameSYNACK:
		return "SYNACK"
	case frameRST:
		return "RST"
	case frameDATA:
		return "DATA"
	case frameFIN:
		return "FIN"
	case framePING:
		return "PING"
	case framePONG:
		return "PONG"
	case frameLISTEN:
		return "LISTEN"
	case frameUNLISTEN:
		return "UNLISTEN"
	case frameCTLRESULT:
		return "CTLRESULT"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", byte(t))
	}
}

const (
	// frameHeaderLen 是固定帧头：[len:4][type:1][streamID:4]。
	frameHeaderLen = 9
	// maxFramePayload 限制单帧 payload，避免恶意对端用超大长度字段骗我们分配内存。
	maxFramePayload = 64 * 1024
)

var errFrameTooLarge = errors.New("frame payload too large")

// frame 是一帧的完整表示。
type frame struct {
	typ      frameType
	streamID uint32
	payload  []byte
}

// encode 把帧序列化成待发送的字节。
//
// 调用方必须持有 Session 的写锁：一次 Write 必须写出完整的一帧，
// 否则并发写会把两帧咬成两个半帧，对端解析会直接错乱。
func (f frame) encode() []byte {
	buf := make([]byte, frameHeaderLen+len(f.payload))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(f.payload)))
	buf[4] = byte(f.typ)
	binary.BigEndian.PutUint32(buf[5:9], f.streamID)
	copy(buf[frameHeaderLen:], f.payload)
	return buf
}

func writeFrame(w io.Writer, f frame) error {
	_, err := w.Write(f.encode())
	return err
}

// readFrame 读一帧：先读满 9 字节头，再按头里的长度读 payload。
//
// 和 v1 握手同一个套路：长度必须写在最前面，否则对端不知道该读到哪儿停。
func readFrame(r io.Reader) (frame, error) {
	head := make([]byte, frameHeaderLen)
	if _, err := io.ReadFull(r, head); err != nil {
		return frame{}, err
	}
	n := int(binary.BigEndian.Uint32(head[0:4]))
	if n > maxFramePayload {
		return frame{}, fmt.Errorf("%w: %d bytes", errFrameTooLarge, n)
	}
	payload := make([]byte, n)
	if _, err := io.ReadFull(r, payload); err != nil {
		return frame{}, err
	}
	return frame{
		typ:      frameType(head[4]),
		streamID: binary.BigEndian.Uint32(head[5:9]),
		payload:  payload,
	}, nil
}
