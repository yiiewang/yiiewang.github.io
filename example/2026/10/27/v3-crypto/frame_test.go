package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	in := frame{typ: frameDATA, streamID: 7, payload: []byte("hello mux")}

	buf := new(bytes.Buffer)
	if err := writeFrame(buf, in); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	out, err := readFrame(buf)
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if out.typ != in.typ || out.streamID != in.streamID || string(out.payload) != string(in.payload) {
		t.Fatalf("frame mismatch: got %v/%d/%q", out.typ, out.streamID, out.payload)
	}
}

func TestFrameRoundTripWithEmptyPayload(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := writeFrame(buf, frame{typ: frameSYNACK, streamID: 2}); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	out, err := readFrame(buf)
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}
	if out.typ != frameSYNACK || out.streamID != 2 || len(out.payload) != 0 {
		t.Fatalf("unexpected frame: %v/%d/%d bytes", out.typ, out.streamID, len(out.payload))
	}
}

// TestReadFrameRejectsTooLargePayload 防止对端用一个巨大的长度字段骗我们分配内存。
func TestReadFrameRejectsTooLargePayload(t *testing.T) {
	head := make([]byte, frameHeaderLen)
	binary.BigEndian.PutUint32(head[0:4], uint32(maxFramePayload+1))
	head[4] = byte(frameDATA)

	if _, err := readFrame(bytes.NewReader(head)); !errors.Is(err, errFrameTooLarge) {
		t.Fatalf("expected errFrameTooLarge, got %v", err)
	}
}

func TestFrameTypeString(t *testing.T) {
	cases := map[frameType]string{
		frameSYN:      "SYN",
		frameSYNACK:   "SYNACK",
		frameRST:      "RST",
		frameDATA:     "DATA",
		frameFIN:      "FIN",
		frameType(99): "UNKNOWN(99)",
	}
	for typ, want := range cases {
		if got := typ.String(); got != want {
			t.Fatalf("frameType %d string: got %q, want %q", byte(typ), got, want)
		}
	}
}
