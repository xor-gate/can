// Copyright 2017 can authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
//
// +build linux

package socketcan

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/xor-gate/can/canio"
)

func TestNew(t *testing.T) {
	c, err := New("vcan0")
	assert.Nil(t, err)
	c.Close()
}

func TestInterfaces(t *testing.T) {
	ifaces, err := Interfaces()
	assert.Nil(t, err)
	assert.True(t, len(ifaces) > 0)
	assert.Contains(t, ifaces, "vcan0")
}

func TestSendRecvLoopbackSFFEmptyFrame(t *testing.T) {
	c, err := New("vcan0")
	assert.Nil(t, err)
	assert.Nil(t, c.Loopback(true))

	c.Send(&canio.Frame{Type: canio.SFF, Id: 0x7ff})

	m, err := c.Recv()
	assert.Nil(t, err)
	assert.Equal(t, canio.SFF, m.Type)
	assert.Equal(t, uint32(0x7ff), m.Id)
	assert.Zero(t, len(m.Data))

	assert.Nil(t, c.Close())
}

func TestSendRecvLoopbackSFFFullFrame(t *testing.T) {
	c, err := New("vcan0")
	assert.Nil(t, err)
	assert.Nil(t, c.Loopback(true))

	assert.Nil(t, c.Send(&canio.Frame{Type: canio.SFF, Id: 0x7ff, Data: []byte("\xde\xad\xbe\xef\xca\xfe\xba\xbe")}))

	m, err := c.Recv()
	assert.Nil(t, err)
	assert.Equal(t, canio.SFF, m.Type)
	assert.Equal(t, uint32(0x7ff), m.Id)
	assert.Equal(t, []byte("\xde\xad\xbe\xef\xca\xfe\xba\xbe"), m.Data)

	assert.Nil(t, c.Close())
}

func TestSendRecvLoopbackEFFEmptyFrame(t *testing.T) {
	c, err := New("vcan0")
	assert.Nil(t, err)
	assert.Nil(t, c.Loopback(true))

	assert.Nil(t, c.Send(&canio.Frame{Type: canio.EFF, Id: 0x1fffffff}))

	m, err := c.Recv()
	assert.Nil(t, err)
	assert.Equal(t, canio.EFF, m.Type)
	assert.Equal(t, uint32(0x1fffffff), m.Id)
	assert.Zero(t, len(m.Data))

	assert.Nil(t, c.Close())
}

func TestSendRecvLoopbackEFFFullFrame(t *testing.T) {
	c, err := New("vcan0")
	assert.Nil(t, err)
	assert.Nil(t, c.Loopback(true))

	assert.Nil(t, c.Send(&canio.Frame{Type: canio.EFF, Id: 0x1fffffff, Data: []byte("\xde\xad\xbe\xef\xca\xfe\xba\xbe")}))

	m, err := c.Recv()
	assert.Nil(t, err)
	assert.Equal(t, canio.EFF, m.Type)
	assert.Equal(t, uint32(0x1fffffff), m.Id)
	assert.Equal(t, []byte("\xde\xad\xbe\xef\xca\xfe\xba\xbe"), m.Data)

	assert.Nil(t, c.Close())
}
