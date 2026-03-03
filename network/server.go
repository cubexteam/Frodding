package network

import (
	"fmt"
	"sync"

	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/raknet"
	"github.com/cubexteam/Frodding/resources"
)

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Debugf(format string, args ...any)
	Chat(name, message string)
}

type Server struct {
	cfg     *resources.Config
	log     Logger
	rak     *raknet.Server
	mu      sync.RWMutex
	sessions map[int64]*Session

	onShutdown func()
}

func NewServer(cfg *resources.Config, log Logger) *Server {
	return &Server{
		cfg:      cfg,
		log:      log,
		sessions: make(map[int64]*Session),
	}
}

func (s *Server) SetShutdownHook(fn func()) {
	s.onShutdown = fn
}

func (s *Server) Start() error {
	guid := int64(0x526f6464696e6721)

	motdFunc := func() string {
		protocol := info.LatestProtocol
		versionTag := info.LatestGameVersionNetwork
		if s.cfg.MotdProtocol > 0 {
			protocol = s.cfg.MotdProtocol
			if vt, ok := info.SupportedProtocols[int32(protocol)]; ok {
				versionTag = vt
			}
		}
		return fmt.Sprintf("MCPE;%s;%d;%s;%d;%d;%d;%s;Survival",
			s.cfg.ServerName, protocol, versionTag,
			s.OnlineCount(), s.cfg.MaxPlayers,
			guid, s.cfg.ServerName,
		)
	}

	s.rak = raknet.NewServer(guid, motdFunc, s)
	s.log.Infof("Starting on %s:%d (MCBE %s, protocol %d)",
		s.cfg.ServerIP, s.cfg.ServerPort,
		info.LatestGameVersionNetwork, info.LatestProtocol)
	err := s.rak.Start(s.cfg.ServerIP, s.cfg.ServerPort)
	if err != nil {
		return err
	}
	s.log.Infof("RakNet server started!")
	return nil
}

func (s *Server) Shutdown() {
	s.rak.Shutdown()
	if s.onShutdown != nil {
		s.onShutdown()
	}
}

func (s *Server) OnConnect(rs *raknet.Session) {
	s.log.Debugf("[NET] New connection guid=%d from %s", rs.GUID, rs.Addr)
	sess := NewSession(s, rs)
	s.mu.Lock()
	s.sessions[rs.GUID] = sess
	s.mu.Unlock()
}

func (s *Server) OnPacket(rs *raknet.Session, payload []byte) {
	s.mu.RLock()
	sess, ok := s.sessions[rs.GUID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	sess.HandlePayload(payload)
}

func (s *Server) OnDisconnect(rs *raknet.Session) {
	s.mu.Lock()
	sess, ok := s.sessions[rs.GUID]
	delete(s.sessions, rs.GUID)
	s.mu.Unlock()
	if ok && sess.Spawned {
		s.BroadcastMessage(fmt.Sprintf("§e%s left the game", sess.Username))
		s.log.Infof("%s left the game", sess.Username)
	}
}

func (s *Server) BroadcastMessage(msg string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sess := range s.sessions {
		if sess.Spawned {
			sess.SendMessage(msg)
		}
	}
}

func (s *Server) OnlineCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

func (s *Server) GetSessions() []*Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Session, 0, len(s.sessions))
	for _, sess := range s.sessions {
		list = append(list, sess)
	}
	return list
}
