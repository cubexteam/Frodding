package raknet

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"strconv"
)

type Stream struct {
	buf []byte
	pos int
}

func NewStream() *Stream {
	return &Stream{}
}

func NewStreamBytes(b []byte) *Stream {
	c := make([]byte, len(b))
	copy(c, b)
	return &Stream{buf: c}
}

func (s *Stream) Bytes() []byte {
	return s.buf
}

func (s *Stream) Len() int {
	return len(s.buf) - s.pos
}

func (s *Stream) Reset() {
	s.buf = s.buf[:0]
	s.pos = 0
}

func (s *Stream) Skip(n int) {
	s.pos += n
}

func (s *Stream) ReadByte() (byte, error) {
	if s.pos >= len(s.buf) {
		return 0, errors.New("stream underflow")
	}
	b := s.buf[s.pos]
	s.pos++
	return b, nil
}

func (s *Stream) ReadBytes(n int) ([]byte, error) {
	if s.pos+n > len(s.buf) {
		return nil, fmt.Errorf("stream underflow: need %d have %d", n, len(s.buf)-s.pos)
	}
	b := s.buf[s.pos : s.pos+n]
	s.pos += n
	return b, nil
}

func (s *Stream) ReadUint16BE() (uint16, error) {
	b, err := s.ReadBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func (s *Stream) ReadUint32BE() (uint32, error) {
	b, err := s.ReadBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

func (s *Stream) ReadInt32BE() (int32, error) {
	v, err := s.ReadUint32BE()
	return int32(v), err
}

func (s *Stream) ReadInt64BE() (int64, error) {
	b, err := s.ReadBytes(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(b)), nil
}

func (s *Stream) ReadUint24LE() (uint32, error) {
	b, err := s.ReadBytes(3)
	if err != nil {
		return 0, err
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16, nil
}

func (s *Stream) ReadBool() (bool, error) {
	b, err := s.ReadByte()
	return b != 0, err
}

func (s *Stream) ReadString() (string, error) {
	n, err := s.ReadUint16BE()
	if err != nil {
		return "", err
	}
	b, err := s.ReadBytes(int(n))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *Stream) ReadMagic() (bool, error) {
	b, err := s.ReadBytes(16)
	if err != nil {
		return false, err
	}
	expected := RaknetMagic()
	for i, v := range expected {
		if b[i] != v {
			return false, nil
		}
	}
	return true, nil
}

func (s *Stream) ReadAddress() (net.UDPAddr, error) {
	ver, err := s.ReadByte()
	if err != nil {
		return net.UDPAddr{}, err
	}
	if ver == 4 {
		ip := make([]byte, 4)
		for i := 0; i < 4; i++ {
			b, err := s.ReadByte()
			if err != nil {
				return net.UDPAddr{}, err
			}
			ip[i] = ^b
		}
		port, err := s.ReadUint16BE()
		if err != nil {
			return net.UDPAddr{}, err
		}
		return net.UDPAddr{IP: net.IP(ip), Port: int(port)}, nil
	}
	b, err := s.ReadBytes(18)
	if err != nil {
		return net.UDPAddr{}, err
	}
	port := binary.BigEndian.Uint16(b[0:2])
	ip := net.IP(b[2:18])
	return net.UDPAddr{IP: ip, Port: int(port)}, nil
}

func (s *Stream) WriteByte(v byte) {
	s.buf = append(s.buf, v)
}

func (s *Stream) WriteBytes(v []byte) {
	s.buf = append(s.buf, v...)
}

func (s *Stream) WriteUint16BE(v uint16) {
	s.buf = append(s.buf, byte(v>>8), byte(v))
}

func (s *Stream) WriteUint32BE(v uint32) {
	s.buf = append(s.buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (s *Stream) WriteInt32BE(v int32) {
	s.WriteUint32BE(uint32(v))
}

func (s *Stream) WriteInt64BE(v int64) {
	s.buf = append(s.buf,
		byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v),
	)
}

func (s *Stream) WriteUint24LE(v uint32) {
	s.buf = append(s.buf, byte(v), byte(v>>8), byte(v>>16))
}

func (s *Stream) WriteBool(v bool) {
	if v {
		s.WriteByte(1)
	} else {
		s.WriteByte(0)
	}
}

func (s *Stream) WriteString(v string) {
	b := []byte(v)
	s.WriteUint16BE(uint16(len(b)))
	s.WriteBytes(b)
}

func (s *Stream) WriteMagic() {
	s.WriteBytes(RaknetMagic())
}

func (s *Stream) WriteAddress(addr net.UDPAddr) {
	ip := addr.IP.To4()
	if ip != nil {
		s.WriteByte(4)
		for _, b := range ip {
			s.WriteByte(^b)
		}
		s.WriteUint16BE(uint16(addr.Port))
	} else {
		s.WriteByte(6)
		s.WriteUint16BE(uint16(addr.Port))
		s.WriteUint32BE(0)
		s.WriteBytes(addr.IP.To16())
		s.WriteUint32BE(0)
	}
}

func (s *Stream) WriteAddressRaw(ipStr string, port uint16) {
	if strings.Contains(ipStr, ".") {
		parts := strings.Split(ipStr, ".")
		s.WriteByte(4)
		for _, p := range parts {
			v, _ := strconv.Atoi(p)
			s.WriteByte(^byte(v))
		}
		s.WriteUint16BE(port)
	} else {
		ip := net.ParseIP(ipStr)
		s.WriteByte(6)
		s.WriteUint16BE(port)
		s.WriteUint32BE(0)
		s.WriteBytes(ip.To16())
		s.WriteUint32BE(0)
	}
}
