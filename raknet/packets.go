package raknet

import (
	"encoding/binary"
	"net"
)

func BuildUnconnectedPong(timestamp int64, serverGUID int64, motd string) []byte {
	s := NewStream()
	s.WriteByte(0x1c)
	s.WriteInt64BE(timestamp)
	s.WriteInt64BE(serverGUID)
	s.WriteMagic()
	s.WriteString(motd)
	return s.Bytes()
}

func BuildOpenConnectionReply1(serverGUID int64, mtu uint16) []byte {
	s := NewStream()
	s.WriteByte(0x06)
	s.WriteMagic()
	s.WriteInt64BE(serverGUID)
	s.WriteBool(false)
	s.WriteUint16BE(mtu)
	return s.Bytes()
}

func BuildOpenConnectionReply2(serverGUID int64, clientAddr net.UDPAddr, mtu uint16) []byte {
	s := NewStream()
	s.WriteByte(0x08)
	s.WriteMagic()
	s.WriteInt64BE(serverGUID)
	s.WriteAddress(clientAddr)
	s.WriteUint16BE(mtu)
	s.WriteBool(false)
	return s.Bytes()
}

func BuildConnectionRequestAccepted(clientAddr net.UDPAddr, clientTimestamp int64, serverTimestamp int64) []byte {
	s := NewStream()
	s.WriteByte(0x10)
	s.WriteAddress(clientAddr)
	s.WriteUint16BE(0)
	for i := 0; i < 10; i++ {
		s.WriteAddress(net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 0})
	}
	s.WriteInt64BE(clientTimestamp)
	s.WriteInt64BE(serverTimestamp)
	return s.Bytes()
}

func BuildConnectedPong(pingTime int64, pongTime int64) []byte {
	s := NewStream()
	s.WriteByte(0x03)
	s.WriteInt64BE(pingTime)
	s.WriteInt64BE(pongTime)
	return s.Bytes()
}

func ParseUnconnectedPing(data []byte) (timestamp int64, guid int64, ok bool) {
	if len(data) < 1+8+16 {
		return
	}
	timestamp = int64(binary.BigEndian.Uint64(data[1:9]))
	magic := RaknetMagic()
	for i, b := range magic {
		if data[9+i] != b {
			return
		}
	}
	if len(data) >= 1+8+16+8 {
		guid = int64(binary.BigEndian.Uint64(data[25:33]))
	}
	ok = true
	return
}

func ParseOCR1(data []byte) (proto byte, mtu uint16, ok bool) {
	if len(data) < 1+16+1 {
		return
	}
	magic := RaknetMagic()
	for i, b := range magic {
		if data[1+i] != b {
			return
		}
	}
	proto = data[17]
	mtu = uint16(len(data) + 28)
	ok = true
	return
}

func ParseOCR2(data []byte) (clientGUID int64, clientAddr net.UDPAddr, mtu uint16, ok bool) {
	s := NewStreamBytes(data)
	s.Skip(1)
	magic := RaknetMagic()
	mb, err := s.ReadBytes(16)
	if err != nil {
		return
	}
	for i, b := range magic {
		if mb[i] != b {
			return
		}
	}
	addr, err := s.ReadAddress()
	if err != nil {
		return
	}
	clientAddr = addr
	mtuV, err := s.ReadUint16BE()
	if err != nil {
		return
	}
	mtu = mtuV
	guidV, err := s.ReadInt64BE()
	if err != nil {
		return
	}
	clientGUID = guidV
	ok = true
	return
}

func ParseConnectionRequest(payload []byte) (clientGUID int64, timestamp int64, ok bool) {
	if len(payload) < 17 {
		return
	}
	if payload[0] != 0x09 {
		return
	}
	s := NewStreamBytes(payload[1:])
	g, err := s.ReadInt64BE()
	if err != nil {
		return
	}
	t, err := s.ReadInt64BE()
	if err != nil {
		return
	}
	clientGUID = g
	timestamp = t
	ok = true
	return
}

func IsNewIncomingConnection(payload []byte) bool {
	return len(payload) > 0 && payload[0] == 0x13
}

func IsDisconnect(payload []byte) bool {
	return len(payload) > 0 && payload[0] == 0x15
}

func ParseConnectedPing(payload []byte) (timestamp int64, ok bool) {
	if len(payload) < 9 || payload[0] != 0x00 {
		return
	}
	s := NewStreamBytes(payload[1:])
	t, err := s.ReadInt64BE()
	if err != nil {
		return
	}
	return t, true
}
