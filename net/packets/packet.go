package packets

import (
	"bytes"
	"encoding/binary"
	"math"
)

type IPacket interface {
	GetID() byte
	Encode()
	Decode()
	GetBuffer() []byte
	SetBuffer([]byte)
}

type Packet struct {
	id  byte
	buf *bytes.Buffer
	raw []byte
	pos int
}

func NewPacket(id byte) *Packet {
	return &Packet{id: id, buf: &bytes.Buffer{}}
}

func (p *Packet) GetID() byte        { return p.id }
func (p *Packet) GetBuffer() []byte  { return p.buf.Bytes() }
func (p *Packet) SetBuffer(data []byte) { p.raw = data; p.pos = 0 }

func (p *Packet) PutByte(v byte)    { p.buf.WriteByte(v) }
func (p *Packet) PutBytes(v []byte) { p.buf.Write(v) }

func (p *Packet) PutBool(v bool) {
	if v {
		p.buf.WriteByte(0x01)
	} else {
		p.buf.WriteByte(0x00)
	}
}

func (p *Packet) PutShort(v int16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	p.buf.Write(b)
}

func (p *Packet) PutLShort(v uint16) {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	p.buf.Write(b)
}

func (p *Packet) PutInt(v int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	p.buf.Write(b)
}

func (p *Packet) PutLInt(v int32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	p.buf.Write(b)
}

func (p *Packet) PutLLong(v int64) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	p.buf.Write(b)
}

func (p *Packet) PutLFloat(v float32) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(v))
	p.buf.Write(b)
}

func (p *Packet) PutVarInt(v int32) {
	p.PutUVarInt((uint32(v) << 1) ^ uint32(v>>31))
}

func (p *Packet) PutUVarInt(v uint32) {
	for v >= 0x80 {
		p.buf.WriteByte(byte(v&0x7f) | 0x80)
		v >>= 7
	}
	p.buf.WriteByte(byte(v))
}

func (p *Packet) PutVarLong(v int64) {
	p.PutUVarLong((uint64(v) << 1) ^ uint64(v>>63))
}

func (p *Packet) PutUVarLong(v uint64) {
	for v >= 0x80 {
		p.buf.WriteByte(byte(v&0x7f) | 0x80)
		v >>= 7
	}
	p.buf.WriteByte(byte(v))
}

func (p *Packet) PutString(v string) {
	p.PutUVarInt(uint32(len(v)))
	p.buf.WriteString(v)
}

func (p *Packet) checkRead(n int) bool { return p.pos+n <= len(p.raw) }

func (p *Packet) GetByte() byte {
	if !p.checkRead(1) { return 0 }
	b := p.raw[p.pos]; p.pos++; return b
}

func (p *Packet) GetBytes(n int) []byte {
	if !p.checkRead(n) { return nil }
	b := p.raw[p.pos : p.pos+n]; p.pos += n; return b
}

func (p *Packet) GetBool() bool { return p.GetByte() != 0x00 }

func (p *Packet) GetShort() int16 {
	b := p.GetBytes(2)
	if b == nil { return 0 }
	return int16(binary.BigEndian.Uint16(b))
}

func (p *Packet) GetLShort() uint16 {
	b := p.GetBytes(2)
	if b == nil { return 0 }
	return binary.LittleEndian.Uint16(b)
}

func (p *Packet) GetInt() int32 {
	b := p.GetBytes(4)
	if b == nil { return 0 }
	return int32(binary.BigEndian.Uint32(b))
}

func (p *Packet) GetLInt() int32 {
	b := p.GetBytes(4)
	if b == nil { return 0 }
	return int32(binary.LittleEndian.Uint32(b))
}

func (p *Packet) GetLLong() int64 {
	b := p.GetBytes(8)
	if b == nil { return 0 }
	return int64(binary.LittleEndian.Uint64(b))
}

func (p *Packet) GetLFloat() float32 {
	b := p.GetBytes(4)
	if b == nil { return 0 }
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

func (p *Packet) GetVarInt() int32 {
	uv := p.GetUVarInt()
	return int32((uv >> 1) ^ -(uv & 1))
}

func (p *Packet) GetUVarInt() uint32 {
	var v uint32
	for i := uint(0); i < 35; i += 7 {
		b := p.GetByte()
		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 { break }
	}
	return v
}

func (p *Packet) GetVarLong() int64 {
	uv := p.GetUVarLong()
	return int64((uv >> 1) ^ -(uv & 1))
}

func (p *Packet) GetUVarLong() uint64 {
	var v uint64
	for i := uint(0); i < 70; i += 7 {
		b := p.GetByte()
		v |= uint64(b&0x7f) << i
		if b&0x80 == 0 { break }
	}
	return v
}

func (p *Packet) GetString() string {
	return string(p.GetBytes(int(p.GetUVarInt())))
}

func (p *Packet) Remaining() int {
	if p.pos >= len(p.raw) { return 0 }
	return len(p.raw) - p.pos
}
