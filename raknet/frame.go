package raknet

import "fmt"

type Frame struct {
	Reliability  Reliability
	Split        bool
	MessageIndex uint32
	OrderIndex   uint32
	OrderChannel byte
	SplitCount   uint32
	SplitID      uint16
	SplitIndex   uint32
	Payload      []byte
}

func (f *Frame) Encode(s *Stream) {
	flags := byte(f.Reliability) << 5
	if f.Split {
		flags |= 0x10
	}
	s.WriteByte(flags)
	s.WriteUint16BE(uint16(len(f.Payload) * 8))

	if f.Reliability.IsReliable() {
		s.WriteUint24LE(f.MessageIndex)
	}
	if f.Reliability.IsOrdered() || f.Reliability.IsSequenced() {
		s.WriteUint24LE(f.OrderIndex)
		s.WriteByte(f.OrderChannel)
	}
	if f.Split {
		s.WriteUint32BE(f.SplitCount)
		s.WriteUint16BE(f.SplitID)
		s.WriteUint32BE(f.SplitIndex)
	}
	s.WriteBytes(f.Payload)
}

func DecodeFrame(s *Stream) (*Frame, error) {
	flags, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	f := &Frame{}
	f.Reliability = Reliability((flags >> 5) & 0x07)
	f.Split = (flags & 0x10) != 0

	bitLen, err := s.ReadUint16BE()
	if err != nil {
		return nil, err
	}
	length := int(bitLen / 8)

	if f.Reliability.IsReliable() {
		f.MessageIndex, err = s.ReadUint24LE()
		if err != nil {
			return nil, err
		}
	}
	if f.Reliability.IsOrdered() || f.Reliability.IsSequenced() {
		f.OrderIndex, err = s.ReadUint24LE()
		if err != nil {
			return nil, err
		}
		f.OrderChannel, err = s.ReadByte()
		if err != nil {
			return nil, err
		}
	}
	if f.Split {
		sc, err := s.ReadUint32BE()
		if err != nil {
			return nil, err
		}
		f.SplitCount = sc
		si, err := s.ReadUint16BE()
		if err != nil {
			return nil, err
		}
		f.SplitID = si
		idx, err := s.ReadUint32BE()
		if err != nil {
			return nil, err
		}
		f.SplitIndex = idx
	}

	f.Payload, err = s.ReadBytes(length)
	if err != nil {
		return nil, fmt.Errorf("frame payload: need %d, %w", length, err)
	}
	return f, nil
}

func (f *Frame) Size() int {
	size := 3
	if f.Reliability.IsReliable() {
		size += 3
	}
	if f.Reliability.IsOrdered() || f.Reliability.IsSequenced() {
		size += 4
	}
	if f.Split {
		size += 10
	}
	size += len(f.Payload)
	return size
}
