package raknet

type Datagram struct {
	SeqNum uint32
	Frames []*Frame
}

func (d *Datagram) Encode() []byte {
	s := NewStream()
	s.WriteByte(0x84)
	s.WriteUint24LE(d.SeqNum)
	for _, f := range d.Frames {
		f.Encode(s)
	}
	return s.Bytes()
}

func DecodeDatagram(data []byte) (*Datagram, error) {
	s := NewStreamBytes(data)
	_, err := s.ReadByte()
	if err != nil {
		return nil, err
	}
	seq, err := s.ReadUint24LE()
	if err != nil {
		return nil, err
	}
	d := &Datagram{SeqNum: seq}
	for s.Len() >= 3 {
		f, err := DecodeFrame(s)
		if err != nil {
			break
		}
		d.Frames = append(d.Frames, f)
	}
	return d, nil
}

func (d *Datagram) Size() int {
	size := 4
	for _, f := range d.Frames {
		size += f.Size()
	}
	return size
}

func EncodeACK(seqNums []uint32) []byte {
	return encodeAckNack(0xc0, seqNums)
}

func EncodeNACK(seqNums []uint32) []byte {
	return encodeAckNack(0xa0, seqNums)
}

func encodeAckNack(id byte, seqNums []uint32) []byte {
	if len(seqNums) == 0 {
		s := NewStream()
		s.WriteByte(id)
		s.WriteUint16BE(0)
		return s.Bytes()
	}

	type run struct{ start, end uint32 }
	var runs []run
	cur := run{seqNums[0], seqNums[0]}
	for _, n := range seqNums[1:] {
		if n == cur.end+1 {
			cur.end = n
		} else {
			runs = append(runs, cur)
			cur = run{n, n}
		}
	}
	runs = append(runs, cur)

	s := NewStream()
	s.WriteByte(id)
	s.WriteUint16BE(uint16(len(runs)))
	for _, r := range runs {
		if r.start == r.end {
			s.WriteBool(true)
			s.WriteUint24LE(r.start)
		} else {
			s.WriteBool(false)
			s.WriteUint24LE(r.start)
			s.WriteUint24LE(r.end)
		}
	}
	return s.Bytes()
}

func DecodeACKNACK(data []byte) ([]uint32, error) {
	s := NewStreamBytes(data)
	s.Skip(1)
	count, err := s.ReadUint16BE()
	if err != nil {
		return nil, err
	}
	var nums []uint32
	for i := 0; i < int(count); i++ {
		single, err := s.ReadBool()
		if err != nil {
			return nil, err
		}
		start, err := s.ReadUint24LE()
		if err != nil {
			return nil, err
		}
		if single {
			nums = append(nums, start)
		} else {
			end, err := s.ReadUint24LE()
			if err != nil {
				return nil, err
			}
			for n := start; n <= end; n++ {
				nums = append(nums, n)
			}
		}
	}
	return nums, nil
}
