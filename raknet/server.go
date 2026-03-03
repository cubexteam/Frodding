package raknet

import (
	"net"
	"sync"
	"time"
)

type Handler interface {
	OnConnect(s *Session)
	OnPacket(s *Session, payload []byte)
	OnDisconnect(s *Session)
}

type Server struct {
	mu       sync.RWMutex
	conn     *net.UDPConn
	sessions map[string]*Session
	guid     int64
	motdFunc func() string
	handler  Handler
	running  bool
}

func NewServer(guid int64, motdFunc func() string, handler Handler) *Server {
	return &Server{
		guid:     guid,
		motdFunc: motdFunc,
		handler:  handler,
		sessions: make(map[string]*Session),
	}
}

func (s *Server) Start(ip string, port int) error {
	addr := &net.UDPAddr{IP: net.ParseIP(ip), Port: port}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	s.conn = conn
	s.running = true
	go s.readLoop()
	go s.tickLoop()
	return nil
}

func (s *Server) Shutdown() {
	s.running = false
	if s.conn != nil {
		s.conn.Close()
	}
}

func (s *Server) readLoop() {
	buf := make([]byte, 2048)
	for s.running {
		n, addr, err := s.conn.ReadFromUDP(buf)
		if err != nil {
			if !s.running {
				return
			}
			continue
		}
		data := make([]byte, n)
		copy(data, buf[:n])
		go func(a *net.UDPAddr, d []byte) {
			defer func() { recover() }()
			s.handlePacket(a, d)
		}(addr, data)
	}
}

func (s *Server) tickLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for s.running {
		<-ticker.C
		s.tick()
	}
}

func (s *Server) tick() {
	s.mu.RLock()
	sessions := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		sessions = append(sessions, sess)
	}
	s.mu.RUnlock()

	for _, sess := range sessions {
		sess.FlushACKs()
		if sess.TimedOut() {
			s.removeSession(sess)
			if s.handler != nil {
				s.handler.OnDisconnect(sess)
			}
		}
	}
}

func (s *Server) handlePacket(addr *net.UDPAddr, data []byte) {
	if len(data) == 0 {
		return
	}
	id := data[0]

	switch {
	case id == 0x01 || id == 0x02:
		ts, _, ok := ParseUnconnectedPing(data)
		if !ok {
			return
		}
		pong := BuildUnconnectedPong(ts, s.guid, s.motdFunc())
		s.conn.WriteToUDP(pong, addr)

	case id == 0x05:
		proto, mtu, ok := ParseOCR1(data)
		if !ok || proto != 8 {
			return
		}
		if mtu > MaxMTU {
			mtu = MaxMTU
		}
		s.conn.WriteToUDP(BuildOpenConnectionReply1(s.guid, mtu), addr)

	case id == 0x07:
		clientGUID, _, mtu, ok := ParseOCR2(data)
		if !ok {
			return
		}
		if mtu > MaxMTU {
			mtu = MaxMTU
		}
		key := addr.String()
		s.mu.Lock()
		if _, exists := s.sessions[key]; exists {
			s.mu.Unlock()
			return
		}
		sess := NewSession(s.conn, addr, clientGUID, mtu)
		s.sessions[key] = sess
		s.mu.Unlock()
		s.conn.WriteToUDP(BuildOpenConnectionReply2(s.guid, *addr, mtu), addr)

	case id >= 0x80 && id <= 0x8f:
		sess := s.getSession(addr)
		if sess == nil || sess.IsDisconnected() {
			return
		}
		s.handleDatagram(sess, data)

	case id == 0xc0:
		sess := s.getSession(addr)
		if sess != nil {
			sess.HandleACK(data)
		}

	case id == 0xa0:
		sess := s.getSession(addr)
		if sess != nil {
			sess.HandleNACK(data)
		}
	}
}

func (s *Server) handleDatagram(sess *Session, data []byte) {
	dg, err := DecodeDatagram(data)
	if err != nil {
		return
	}

	sess.mu.Lock()
	sess.lastRecv = time.Now()

	if sess.recvSeqs[dg.SeqNum] {
		sess.mu.Unlock()
		return
	}
	sess.recvSeqs[dg.SeqNum] = true
	sess.pendingACK = append(sess.pendingACK, dg.SeqNum)

	if dg.SeqNum > sess.recvSeq+1 {
		for n := sess.recvSeq + 1; n < dg.SeqNum; n++ {
			if !sess.recvSeqs[n] {
				sess.pendingNACK = append(sess.pendingNACK, n)
			}
		}
	}
	if dg.SeqNum >= sess.recvSeq {
		sess.recvSeq = dg.SeqNum + 1
	}
	sess.mu.Unlock()

	for _, f := range dg.Frames {
		s.handleFrame(sess, f)
	}
}

func (s *Server) handleFrame(sess *Session, f *Frame) {
	if f.Split {
		sess.mu.Lock()
		buf, ok := sess.splitMap[f.SplitID]
		if !ok {
			buf = &splitBuf{count: f.SplitCount, parts: make(map[uint32][]byte)}
			sess.splitMap[f.SplitID] = buf
		}
		buf.parts[f.SplitIndex] = f.Payload
		assembled := false
		var full []byte
		if uint32(len(buf.parts)) >= buf.count {
			for i := uint32(0); i < buf.count; i++ {
				full = append(full, buf.parts[i]...)
			}
			delete(sess.splitMap, f.SplitID)
			assembled = true
		}
		sess.mu.Unlock()
		if assembled {
			combined := &Frame{
				Reliability:  f.Reliability,
				MessageIndex: f.MessageIndex,
				OrderIndex:   f.OrderIndex,
				OrderChannel: f.OrderChannel,
				Payload:      full,
			}
			s.handleFrame(sess, combined)
		}
		return
	}

	if f.Reliability.IsOrdered() {
		ch := int(f.OrderChannel)
		sess.mu.Lock()
		expected := sess.orderNext[ch]
		if f.OrderIndex == expected {
			sess.orderNext[ch]++
			payload := f.Payload
			var extras [][]byte
			for {
				next := sess.orderQueues[ch][sess.orderNext[ch]]
				if next == nil {
					break
				}
				delete(sess.orderQueues[ch], sess.orderNext[ch])
				sess.orderNext[ch]++
				extras = append(extras, next)
			}
			sess.mu.Unlock()
			s.dispatch(sess, payload)
			for _, e := range extras {
				s.dispatch(sess, e)
			}
		} else if f.OrderIndex > expected {
			sess.orderQueues[ch][f.OrderIndex] = f.Payload
			sess.mu.Unlock()
		} else {
			sess.mu.Unlock()
		}
		return
	}

	s.dispatch(sess, f.Payload)
}

func (s *Server) dispatch(sess *Session, payload []byte) {
	if len(payload) == 0 {
		return
	}
	id := payload[0]

	if id == 0x09 {
		_, timestamp, ok := ParseConnectionRequest(payload)
		if !ok {
			return
		}
		accepted := BuildConnectionRequestAccepted(*sess.Addr, timestamp, sess.Timestamp())
		sess.Send(accepted, ReliableOrdered, 0)
		return
	}

	if id == 0x13 {
		sess.SetConnected()
		if s.handler != nil {
			s.handler.OnConnect(sess)
		}
		return
	}

	if id == 0x15 {
		s.removeSession(sess)
		if s.handler != nil {
			s.handler.OnDisconnect(sess)
		}
		return
	}

	if id == 0x00 {
		t, ok := ParseConnectedPing(payload)
		if ok {
			pong := BuildConnectedPong(t, sess.Timestamp())
			sess.Send(pong, Unreliable, 0)
		}
		return
	}

	if !sess.IsConnected() {
		return
	}

	if s.handler != nil {
		s.handler.OnPacket(sess, payload)
	}
}

func (s *Server) getSession(addr *net.UDPAddr) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[addr.String()]
}

func (s *Server) removeSession(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sess.Addr.String())
}

func (s *Server) GetSessionByGUID(guid int64) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sess := range s.sessions {
		if sess.GUID == guid {
			return sess
		}
	}
	return nil
}

func (s *Server) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

func (s *Server) Sessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, sess)
	}
	return list
}
