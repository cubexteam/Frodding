package raknet

import (
	"net"
	"sync"
	"time"
)

type SessionState int

const (
	StateHandshaking SessionState = iota
	StateConnected
	StateDisconnected
)

type Session struct {
	mu   sync.Mutex
	Addr *net.UDPAddr
	GUID int64
	MTU  uint16
	conn *net.UDPConn

	state SessionState

	sendSeq  uint32
	recvSeq  uint32
	msgIndex uint32

	orderIndex  [MaxChannels]uint32
	orderNext   [MaxChannels]uint32
	orderQueues [MaxChannels]map[uint32][]byte

	recvSeqs    map[uint32]bool
	pendingACK  []uint32
	pendingNACK []uint32

	splitMap    map[uint16]*splitBuf

	recoveryMap map[uint32]*Datagram

	lastRecv time.Time
	connTime time.Time
}

type splitBuf struct {
	count uint32
	parts map[uint32][]byte
}

func NewSession(conn *net.UDPConn, addr *net.UDPAddr, guid int64, mtu uint16) *Session {
	s := &Session{
		conn:        conn,
		Addr:        addr,
		GUID:        guid,
		MTU:         mtu,
		state:       StateHandshaking,
		recvSeqs:    make(map[uint32]bool),
		splitMap:    make(map[uint16]*splitBuf),
		recoveryMap: make(map[uint32]*Datagram),
		lastRecv:    time.Now(),
	}
	for i := 0; i < MaxChannels; i++ {
		s.orderQueues[i] = make(map[uint32][]byte)
	}
	return s
}

func (s *Session) IsConnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == StateConnected
}

func (s *Session) IsDisconnected() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == StateDisconnected
}

func (s *Session) SetConnected() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateConnected
	s.connTime = time.Now()
}

func (s *Session) Timestamp() int64 {
	if s.connTime.IsZero() {
		return 0
	}
	return time.Since(s.connTime).Milliseconds()
}

func (s *Session) HandleACK(data []byte) {
	nums, err := DecodeACKNACK(data)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range nums {
		delete(s.recoveryMap, n)
	}
}

func (s *Session) HandleNACK(data []byte) {
	nums, err := DecodeACKNACK(data)
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range nums {
		dg, ok := s.recoveryMap[n]
		if ok {
			s.conn.WriteToUDP(dg.Encode(), s.Addr)
		}
	}
}

func (s *Session) FlushACKs() {
	s.mu.Lock()
	acks := s.pendingACK
	nacks := s.pendingNACK
	s.pendingACK = nil
	s.pendingNACK = nil
	s.mu.Unlock()

	if len(acks) > 0 {
		s.conn.WriteToUDP(EncodeACK(acks), s.Addr)
	}
	if len(nacks) > 0 {
		s.conn.WriteToUDP(EncodeNACK(nacks), s.Addr)
	}
}

func (s *Session) Send(payload []byte, reliability Reliability, orderChannel byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxPayload := int(s.MTU) - 4 - 3
	if reliability.IsReliable() {
		maxPayload -= 3
	}
	if reliability.IsOrdered() || reliability.IsSequenced() {
		maxPayload -= 4
	}
	if maxPayload < 100 {
		maxPayload = 100
	}

	if len(payload) <= maxPayload {
		f := &Frame{Reliability: reliability, Payload: payload}
		if reliability.IsReliable() {
			f.MessageIndex = s.msgIndex
			s.msgIndex++
		}
		if reliability.IsOrdered() {
			f.OrderIndex = s.orderIndex[orderChannel]
			f.OrderChannel = orderChannel
			s.orderIndex[orderChannel]++
		}
		s.sendDatagram([]*Frame{f})
		return
	}

	splitID := s.splitID()
	count := (len(payload) + maxPayload - 1) / maxPayload
	orderIdx := s.orderIndex[orderChannel]
	if reliability.IsOrdered() {
		s.orderIndex[orderChannel]++
	}
	for i := 0; i < count; i++ {
		start := i * maxPayload
		end := start + maxPayload
		if end > len(payload) {
			end = len(payload)
		}
		f := &Frame{
			Reliability: reliability,
			Split:       true,
			SplitCount:  uint32(count),
			SplitID:     splitID,
			SplitIndex:  uint32(i),
			Payload:     make([]byte, end-start),
		}
		copy(f.Payload, payload[start:end])
		if reliability.IsReliable() {
			f.MessageIndex = s.msgIndex
			s.msgIndex++
		}
		if reliability.IsOrdered() {
			f.OrderIndex = orderIdx
			f.OrderChannel = orderChannel
		}
		s.sendDatagram([]*Frame{f})
	}
}

func (s *Session) splitID() uint16 {
	id := uint16(s.msgIndex & 0xffff)
	return id
}

func (s *Session) sendDatagram(frames []*Frame) {
	seq := s.sendSeq
	s.sendSeq++
	dg := &Datagram{SeqNum: seq, Frames: frames}
	pkt := dg.Encode()
	s.conn.WriteToUDP(pkt, s.Addr)

	hasReliable := false
	for _, f := range frames {
		if f.Reliability.IsReliable() {
			hasReliable = true
			break
		}
	}
	if hasReliable {
		s.recoveryMap[seq] = dg
	}
}

func (s *Session) SendRaw(data []byte) {
	s.conn.WriteToUDP(data, s.Addr)
}

func (s *Session) TimedOut() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return time.Since(s.lastRecv) > SessionTimeout
}

func (s *Session) Disconnect() {
	s.mu.Lock()
	s.state = StateDisconnected
	seq := s.sendSeq
	s.sendSeq++
	s.mu.Unlock()
	f := &Frame{Reliability: Unreliable, Payload: []byte{0x15}}
	dg := &Datagram{SeqNum: seq, Frames: []*Frame{f}}
	s.conn.WriteToUDP(dg.Encode(), s.Addr)
}
