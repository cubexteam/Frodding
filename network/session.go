package network

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
	"github.com/cubexteam/Frodding/net/packets/bedrock"
	"github.com/cubexteam/Frodding/raknet"
	"github.com/google/uuid"
)

type Session struct {
	mu      sync.Mutex
	server  *Server
	rak     *raknet.Session

	UUID     uuid.UUID
	XUID     string
	Username string
	ClientID int64
	Language string
	SkinID   string
	SkinData []byte
	CapeData []byte
	GeomName string
	GeomData string

	Spawned      bool
	X, Y, Z      float32
	ViewDistance int32
	sentChunks   map[[2]int32]struct{}
}

func NewSession(srv *Server, rak *raknet.Session) *Session {
	return &Session{
		server:       srv,
		rak:          rak,
		UUID:         uuid.New(),
		ViewDistance: 8,
		sentChunks:   make(map[[2]int32]struct{}),
		Y:            64,
	}
}

func (s *Session) SendPacket(pk packets.IPacket) {
	pk.Encode()
	s.sendBatch(pk.GetID(), pk.GetBuffer())
}

func (s *Session) sendBatch(id byte, data []byte) {
	var raw bytes.Buffer
	raw.Grow(1 + len(data) + 5)
	raw.WriteByte(id)
	raw.Write(data)

	var buf bytes.Buffer
	buf.Grow(raw.Len() + 5)
	writeVarUint32(&buf, uint32(raw.Len()))
	buf.Write(raw.Bytes())

	var compressed bytes.Buffer
	w, _ := zlib.NewWriterLevel(&compressed, zlib.BestSpeed)
	_, _ = w.Write(buf.Bytes())
	_ = w.Close()

	batch := make([]byte, 1+compressed.Len())
	batch[0] = 0xfe
	copy(batch[1:], compressed.Bytes())
	s.rak.Send(batch, raknet.ReliableOrdered, 0)
}

func writeVarUint32(buf *bytes.Buffer, v uint32) {
	for v >= 0x80 {
		buf.WriteByte(byte(v&0x7f) | 0x80)
		v >>= 7
	}
	buf.WriteByte(byte(v))
}

func (s *Session) HandlePayload(data []byte) {
	if len(data) == 0 || data[0] != 0xfe {
		return
	}
	r, err := zlib.NewReader(bytes.NewReader(data[1:]))
	if err != nil {
		return
	}
	defer func() { _ = r.Close() }()
	decompressed, err := io.ReadAll(r)
	if err != nil {
		return
	}
	reader := bytes.NewReader(decompressed)
	for reader.Len() > 0 {
		pkLen, err := readVarUint32(reader)
		if err != nil || pkLen == 0 || int(pkLen) > reader.Len() {
			break
		}
		pkData := make([]byte, pkLen)
		_, _ = reader.Read(pkData)
		s.handlePacketData(pkData)
	}
}

func readVarUint32(r *bytes.Reader) (uint32, error) {
	var v uint32
	for i := uint(0); i < 35; i += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		v |= uint32(b&0x7f) << i
		if b&0x80 == 0 {
			break
		}
	}
	return v, nil
}

func (s *Session) handlePacketData(data []byte) {
	if len(data) == 0 {
		return
	}
	id := data[0]
	body := data[1:]

	switch id {
	case info.IDLogin:
		pk := bedrock.NewLoginPacket()
		pk.SetBuffer(body)
		pk.Decode()
		s.handleLogin(pk)
	case info.IDResourcePackClientResponse:
		pk := bedrock.NewResourcePackClientResponsePacket()
		pk.SetBuffer(body)
		pk.Decode()
		s.handleResourcePackResponse(pk)
	case info.IDMovePlayer:
		pk := bedrock.NewMovePlayerPacket()
		pk.SetBuffer(body)
		pk.Decode()
		s.handleMovePlayer(pk)
	case info.IDRequestChunkRadius:
		pk := bedrock.NewRequestChunkRadiusPacket()
		pk.SetBuffer(body)
		pk.Decode()
		s.handleChunkRadius(pk)
	case info.IDText:
		pk := bedrock.NewTextPacket(0, "", "")
		pk.SetBuffer(body)
		pk.Decode()
		s.handleText(pk)
	}
}

func (s *Session) handleLogin(pk *bedrock.LoginPacket) {
	s.server.log.Infof("Player connecting: %s (protocol %d)", pk.Username, pk.Protocol)

	if pk.Protocol < info.MinProtocol {
		s.SendPacket(bedrock.NewPlayStatusPacket(bedrock.PlayStatusLoginFailedClient))
		return
	}
	if pk.Protocol > info.LatestProtocol {
		s.SendPacket(bedrock.NewPlayStatusPacket(bedrock.PlayStatusLoginFailedServer))
		return
	}

	s.UUID = pk.ClientUUID
	s.XUID = pk.ClientXUID
	s.Username = pk.Username
	s.ClientID = pk.ClientID
	s.Language = pk.Language
	s.SkinID = pk.SkinID
	s.SkinData = pk.SkinData
	s.CapeData = pk.CapeData
	s.GeomName = pk.GeometryName
	s.GeomData = pk.GeometryData

	s.SendPacket(bedrock.NewPlayStatusPacket(bedrock.PlayStatusLoginSuccess))
	s.SendPacket(bedrock.NewResourcePackInfoPacket(false))
}

func (s *Session) handleResourcePackResponse(pk *bedrock.ResourcePackClientResponsePacket) {
	switch pk.Status {
	case bedrock.ResourcePackResponseSendPacks:
		s.SendPacket(bedrock.NewResourcePackInfoPacket(false))
	case bedrock.ResourcePackResponseAllPacksSent:
		s.SendPacket(bedrock.NewResourcePackStackPacket(false))
	case bedrock.ResourcePackResponseCompleted:
		s.startGame()
	case bedrock.ResourcePackResponseRefused:
		s.SendPacket(bedrock.NewResourcePackStackPacket(false))
	}
}

func (s *Session) startGame() {
	sg := bedrock.NewStartGamePacket()
	sg.EntityUniqueID = 1
	sg.EntityRuntimeID = 1
	sg.PlayerGameMode = 0
	sg.SpawnX = s.X
	sg.SpawnY = s.Y + 1
	sg.SpawnZ = s.Z
	sg.Seed = 12345
	sg.Dimension = 0
	sg.Generator = 1
	sg.GameMode = 0
	sg.Difficulty = 1
	sg.SpawnBY = 64
	sg.AchievementsDisabled = true
	sg.Time = 6000
	sg.IsMultiplayer = true
	sg.BroadcastToLAN = true
	sg.CommandsEnabled = true
	sg.LevelName = s.server.cfg.ServerName
	sg.CurrentTick = 0
	sg.EnchantmentSeed = 0
	s.SendPacket(sg)

	s.SendPacket(bedrock.NewSetTimePacket(6000))

	adv := bedrock.NewAdventureSettingsPacket()
	adv.Flags = 0x20 | 0x80
	adv.PlayerPermission = 1
	adv.EntityID = 1
	s.SendPacket(adv)

	attr := bedrock.NewUpdateAttributesPacket(1)
	attr.Attributes = bedrock.DefaultAttributes()
	s.SendPacket(attr)
}

func (s *Session) handleChunkRadius(pk *bedrock.RequestChunkRadiusPacket) {
	radius := pk.Radius
	if radius > 4 {
		radius = 4
	} else if radius < 2 {
		radius = 2
	}
	s.ViewDistance = radius
	s.SendPacket(bedrock.NewChunkRadiusUpdatedPacket(radius))
	s.sendChunks()
}

func (s *Session) sendChunks() {
	cx := int32(s.X) >> 4
	cz := int32(s.Z) >> 4
	r := s.ViewDistance
	chunkData := bedrock.BuildFlatChunk()

	s.mu.Lock()
	defer s.mu.Unlock()

	for x := cx - r; x <= cx+r; x++ {
		for z := cz - r; z <= cz+r; z++ {
			key := [2]int32{x, z}
			if _, sent := s.sentChunks[key]; sent {
				continue
			}
			s.sentChunks[key] = struct{}{}
			s.SendPacket(bedrock.NewFullChunkDataPacket(x, z, chunkData))
		}
	}

	if !s.Spawned {
		s.Spawned = true
		s.SendPacket(bedrock.NewPlayStatusPacket(bedrock.PlayStatusPlayerSpawn))
		s.server.log.Infof("%s joined the game", s.Username)
		s.server.BroadcastMessage(fmt.Sprintf("§e%s joined the game", s.Username))
	}
}

func (s *Session) handleMovePlayer(pk *bedrock.MovePlayerPacket) {
	s.X = pk.X
	s.Y = pk.Y
	s.Z = pk.Z
}

func (s *Session) handleText(pk *bedrock.TextPacket) {
	msg := pk.Message
	if strings.HasPrefix(msg, "/") {
		s.handleCommand(strings.TrimPrefix(msg, "/"))
		return
	}
	s.server.log.Chat(s.Username, msg)
	s.server.BroadcastMessage(fmt.Sprintf("<%s> %s", s.Username, msg))
}

var gameModeNames = map[int32]string{
	0: "Survival",
	1: "Creative",
	2: "Adventure",
}

func (s *Session) handleCommand(cmd string) {
	args := strings.Fields(cmd)
	if len(args) == 0 {
		return
	}
	switch strings.ToLower(args[0]) {
	case "list":
		sessions := s.server.GetSessions()
		names := make([]string, 0, len(sessions))
		for _, sess := range sessions {
			if sess.Spawned {
				names = append(names, sess.Username)
			}
		}
		s.SendMessage(fmt.Sprintf("§eOnline (%d/%d): %s", len(names), s.server.cfg.MaxPlayers, strings.Join(names, ", ")))

	case "gm":
		if len(args) < 2 {
			s.SendMessage("§cUsage: /gm <0|1|2>")
			return
		}
		gm, ok := parseGameMode(args[1])
		if !ok {
			s.SendMessage("§cUsage: /gm <0|1|2>")
			return
		}
		s.setGameMode(gm)

	case "stop":
		s.server.BroadcastMessage("§cServer is shutting down!")
		s.server.Shutdown()

	default:
		s.SendMessage(fmt.Sprintf("§cUnknown command: %s", args[0]))
	}
}

func parseGameMode(s string) (int32, bool) {
	switch s {
	case "0", "s", "survival":
		return 0, true
	case "1", "c", "creative":
		return 1, true
	case "2", "a", "adventure":
		return 2, true
	}
	return 0, false
}

func (s *Session) setGameMode(gm int32) {
	s.SendPacket(bedrock.NewSetGameModePacket(gm))
	adv := bedrock.NewAdventureSettingsPacket()
	if gm == 1 {
		adv.Flags = 0x20 | 0x40
		adv.Flags2 = 0x40
	} else {
		adv.Flags = 0x20
	}
	adv.PlayerPermission = 1
	adv.EntityID = 1
	s.SendPacket(adv)
	name := gameModeNames[gm]
	s.SendMessage(fmt.Sprintf("§aGame mode set to %s", name))
}

func (s *Session) Disconnect(reason string) {
	s.SendPacket(bedrock.NewDisconnectPacket(reason, false))
}

func (s *Session) SendMessage(msg string) {
	s.SendPacket(bedrock.NewTextPacket(bedrock.TextTypeSystem, "", msg))
}
