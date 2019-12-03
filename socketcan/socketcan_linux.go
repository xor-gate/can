// Copyright 2017 can authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package socketcan

import (
	"net"
	"fmt"
	"bytes"
	"errors"
	"unsafe"
	"strings"
	"encoding/binary"
	"golang.org/x/sys/unix"
	"github.com/xor-gate/can/canio"
)

const (
	kCAN_RAW_RECV_OWN_MSGS = 4
	kSOL_CAN_RAW           = 101
)

type SocketCAN struct {
	ifname string
	fd int
}

type msg struct {
	id uint32 // 32 bit CAN_ID + EFF/RTR/ERR flags
	dlc uint8 // frame payload length in byte (0 .. CAN_MAX_DLEN)
	_ [3] byte // padding
	data [8] byte // data
}

type ifreqIndex struct {
	Name  [16]byte
	Index int
}

func ioctlIfreq(fd int, ifreq *ifreqIndex) error {
	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		unix.SIOCGIFINDEX,
		uintptr(unsafe.Pointer(ifreq)),
	)
	if errno != 0 {
		return fmt.Errorf("ioctl: %v", errno)
	}
	return nil
}

func getIfIndex(fd int, ifName string) (int, error) {
	ifNameRaw, err := unix.ByteSliceFromString(ifName)
	if err != nil {
		return 0, err
	}
	if len(ifNameRaw) > 16 {
		return 0, errors.New("maximum ifname length is 16 characters")
	}

	ifReq := ifreqIndex{}
	copy(ifReq.Name[:], ifNameRaw)
	err = ioctlIfreq(fd, &ifReq)
	return ifReq.Index, err
}

// Interfaces returns a list of the system's CAN interface names
func Interfaces() ([]string, error) {
	nifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	ifaces := make([]string, 0)
	for _, v := range nifaces {
		if strings.HasPrefix(v.Name, "can") || strings.HasPrefix(v.Name, "vcan") {
			ifaces = append(ifaces, v.Name)
		}
	}

	return ifaces, nil
}

// New creates a new SocketCAN instance on the ifname
func New(ifname string) (*SocketCAN, error) {
	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		return nil, err
	}

	ifIndex, err := getIfIndex(fd, ifname)
	if err != nil {
		return nil, err
	}

	addr := &unix.SockaddrCAN{Ifindex: ifIndex}
	if err := unix.Bind(fd, addr); err != nil {
		return nil, err
	}

	s := &SocketCAN{fd: fd, ifname: ifname}
	return s, nil
}

// Send will send a single CAN frame
func (s *SocketCAN) Send(m *canio.Frame) error {
	buf := new(bytes.Buffer)
	_m := &msg{id: m.Id, dlc: uint8(len(m.Data))}

	// prepare the id
	switch m.Type {
	case canio.SFF:
		_m.id &= unix.CAN_SFF_MASK
	case canio.EFF:
		_m.id &= unix.CAN_EFF_MASK
		_m.id |= unix.CAN_EFF_FLAG
	case canio.RTR:
		// XXX not sure but probably use the EFF mask
		m.Id &= unix.CAN_EFF_MASK
		_m.id |= unix.CAN_RTR_FLAG
	case canio.ERR:
		m.Id &= unix.CAN_ERR_MASK
		_m.id |= unix.CAN_ERR_FLAG
	}

	// copy data
	if _m.dlc > 0 {
		copy(_m.data[:_m.dlc], m.Data)
	}

	// encode and write to fd
	_ = binary.Write(buf, binary.LittleEndian, _m)
	unix.Write(s.fd, buf.Bytes())

	return nil
}

// Recv receives a single CAN frame
func (s *SocketCAN) Recv() (*canio.Frame, error) {
	// read a single socketcan frame
	f := make([]byte, 16)
	unix.Read(s.fd, f)

	m := &canio.Frame{}
	r := bytes.NewReader(f)

	// canid
	_ = binary.Read(r, binary.LittleEndian, &m.Id)

	if m.Id & unix.CAN_EFF_FLAG != 0 {
		m.Type = canio.EFF
		m.Id &= unix.CAN_EFF_MASK
	} else if m.Id & unix.CAN_ERR_FLAG != 0 {
		m.Type = canio.ERR
		m.Id &= unix.CAN_ERR_MASK
	} else if m.Id & unix.CAN_RTR_FLAG != 0 {
		m.Type = canio.RTR
		// XXX not sure but probably use the EFF mask
		m.Id &= unix.CAN_EFF_MASK
	} else {
		m.Type = canio.SFF
		m.Id &= unix.CAN_SFF_MASK
	}

	// dlc
	var dlc uint8
	_ = binary.Read(r, binary.LittleEndian, &dlc)

	// padding
	pad := make([]byte, 3)
	r.Read(pad)

	// data
	m.Data = make([]byte, dlc)
	r.Read(m.Data)
	return m, nil
}

// Loopback controls the frame loopback mode
func (s *SocketCAN) Loopback(enable bool) error {
	var _enable int
	if enable {
		_enable = 1
	}
	return unix.SetsockoptInt(s.fd, kSOL_CAN_RAW, kCAN_RAW_RECV_OWN_MSGS, _enable)
}

// Close will close the socketcan interface
func (s *SocketCAN) Close() error {
	unix.Close(s.fd)
	s.fd = -1
	return nil
}

// Name returns the interface name
func (s *SocketCAN) Name() string {
	return s.ifname
}
