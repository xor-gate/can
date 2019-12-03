// Copyright 2017 can authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package canio

type FrameType uint

const (
	SFF FrameType = iota // SFF frame format
	EFF // EFF extended frame format
	RTR // RTR frame format
	ERR // ERR frame format
)

// Frame represents a single CAN frame
type Frame struct {
	Type FrameType
	Id uint32
	Data []byte
}
