package packets

import (
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
	buf []byte
	raw []byte
	pos int
}

func NewPacket(id byte) *Packet {
	return &Packet{id: id}
}

func (p *Packet) GetID() byte         { return p.id }
func (p *Packet) GetBuffer() []byte   { return p.buf }
func (p *Packet) SetBuffer(data []byte) { p.raw = data; p.pos = 0 }

func (p *Packet) PutByte(v byte)    { p.buf = append(p.buf, v) }
func (p *Packet) PutBytes(v []byte) { p.buf = append(p.buf, v...) }

func (p *Packet) PutBool(v bool) {
	if v {
		p.buf = append(p.buf, 0x01)
	} else {
		p.buf = append(p.buf, 0x00)
	}
}

func (p *Packet) PutShort(v int16) {
	p.buf = append(p.buf, byte(v>>8), byte(v))
}

func (p *Packet) PutLShort(v uint16) {
	p.buf = append(p.buf, byte(v), byte(v>>8))
}

func (p *Packet) PutInt(v int32) {
	p.buf = append(p.buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func (p *Packet) PutLInt(v int32) {
	u := uint32(v)
	p.buf = append(p.buf, byte(u), byte(u>>8), byte(u>>16), byte(u>>24))
}

func (p *Packet) PutLLong(v int64) {
	u := uint64(v)
	p.buf = append(p.buf,
		byte(u), byte(u>>8), byte(u>>16), byte(u>>24),
		byte(u>>32), byte(u>>40), byte(u>>48), byte(u>>56),
	)
}

func (p *Packet) PutLFloat(v float32) {
	u := math.Float32bits(v)
	p.buf = append(p.buf, byte(u), byte(u>>8), byte(u>>16), byte(u>>24))
}

func (p *Packet) PutVarInt(v int32) {
	p.PutUVarInt((uint32(v) << 1) ^ uint32(v>>31))
}

func (p *Packet) PutUVarInt(v uint32) {
	for v >= 0x80 {
		p.buf = append(p.buf, byte(v&0x7f)|0x80)
		v >>= 7
	}
	p.buf = append(p.buf, byte(v))
}

func (p *Packet) PutVarLong(v int64) {
	p.PutUVarLong((uint64(v) << 1) ^ uint64(v>>63))
}

func (p *Packet) PutUVarLong(v uint64) {
	for v >= 0x80 {
		p.buf = append(p.buf, byte(v&0x7f)|0x80)
		v >>= 7
	}
	p.buf = append(p.buf, byte(v))
}

func (p *Packet) PutString(v string) {
	p.PutUVarInt(uint32(len(v)))
	p.buf = append(p.buf, v...)
}

func (p *Packet) checkRead(n int) bool { return p.pos+n <= len(p.raw) }

func (p *Packet) GetByte() byte {
	if !p.checkRead(1) {
		return 0
	}
	b := p.raw[p.pos]
	p.pos++
	return b
}

func (p *Packet) GetBytes(n int) []byte {
	if !p.checkRead(n) {
		return nil
	}
	b := p.raw[p.pos : p.pos+n]
	p.pos += n
	return b
}

func (p *Packet) GetBool() bool { return p.GetByte() != 0 }

func (p *Packet) GetShort() int16 {
	if b := p.GetBytes(2); b != nil {
		return int16(binary.BigEndian.Uint16(b))
	}
	return 0
}

func (p *Packet) GetLShort() uint16 {
	if b := p.GetBytes(2); b != nil {
		return binary.LittleEndian.Uint16(b)
	}
	return 0
}

func (p *Packet) GetInt() int32 {
	if b := p.GetBytes(4); b != nil {
		return int32(binary.BigEndian.Uint32(b))
	}
	return 0
}

func (p *Packet) GetLInt() int32 {
	if b := p.GetBytes(4); b != nil {
		return int32(binary.LittleEndian.Uint32(b))
	}
	return 0
}

func (p *Packet) GetLLong() int64 {
	if b := p.GetBytes(8); b != nil {
		return int64(binary.LittleEndian.Uint64(b))
	}
	return 0
}

func (p *Packet) GetLFloat() float32 {
	if b := p.GetBytes(4); b != nil {
		return math.Float32frombits(binary.LittleEndian.Uint32(b))
	}
	return 0
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
		if b&0x80 == 0 {
			break
		}
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
		if b&0x80 == 0 {
			break
		}
	}
	return v
}

func (p *Packet) GetString() string {
	n := p.GetUVarInt()
	b := p.GetBytes(int(n))
	return string(b)
}

func (p *Packet) Remaining() int {
	if p.pos >= len(p.raw) {
		return 0
	}
	return len(p.raw) - p.pos
}
