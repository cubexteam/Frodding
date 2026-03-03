package bedrock

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"strings"

	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
	"github.com/google/uuid"
)

type LoginPacket struct {
	*packets.Packet
	Protocol     int32
	Username     string
	ClientUUID   uuid.UUID
	ClientXUID   string
	ClientID     int64
	ServerAddr   string
	Language     string
	SkinID       string
	SkinData     []byte
	CapeData     []byte
	GeometryName string
	GeometryData string
}

func NewLoginPacket() *LoginPacket {
	return &LoginPacket{
		Packet:     packets.NewPacket(info.IDLogin),
		ClientUUID: uuid.New(),
	}
}

func (pk *LoginPacket) Encode() {}

func (pk *LoginPacket) Decode() {
	pk.Protocol = pk.GetInt()
	pk.GetByte()
	totalLen := int(pk.GetUVarInt())
	if totalLen <= 0 || totalLen > pk.Remaining() {
		pk.Username = "Player"
		return
	}
	blob := pk.GetBytes(totalLen)
	if len(blob) < 4 {
		pk.Username = "Player"
		return
	}
	chainLen := int(binary.LittleEndian.Uint32(blob[:4]))
	if chainLen <= 0 || 4+chainLen > len(blob) {
		pk.Username = "Player"
		return
	}
	var chainData struct {
		Chain []string `json:"chain"`
	}
	if err := json.Unmarshal(blob[4:4+chainLen], &chainData); err == nil {
		for _, raw := range chainData.Chain {
			pk.parseChain(raw)
		}
	}
	offset := 4 + chainLen
	if offset+4 > len(blob) {
		if pk.Username == "" {
			pk.Username = "Player"
		}
		return
	}
	clientLen := int(binary.LittleEndian.Uint32(blob[offset : offset+4]))
	offset += 4
	if clientLen > 0 && offset+clientLen <= len(blob) {
		pk.parseClientData(string(blob[offset : offset+clientLen]))
	}
	if pk.Username == "" {
		pk.Username = "Player"
	}
}

func decodeJWTPayload(s string) ([]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		switch len(s) % 4 {
		case 2:
			s += "=="
		case 3:
			s += "="
		}
		b, err = base64.URLEncoding.DecodeString(s)
	}
	return b, err
}

func (pk *LoginPacket) parseChain(raw string) {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return
	}
	payload, err := decodeJWTPayload(parts[1])
	if err != nil {
		return
	}
	var data struct {
		ExtraData map[string]interface{} `json:"extraData"`
	}
	if err := json.Unmarshal(payload, &data); err != nil || data.ExtraData == nil {
		return
	}
	if v, ok := data.ExtraData["displayName"]; ok {
		if name, _ := v.(string); name != "" {
			pk.Username = name
		}
	}
	if v, ok := data.ExtraData["identity"]; ok {
		if idStr, _ := v.(string); idStr != "" {
			if u, err := uuid.Parse(idStr); err == nil {
				pk.ClientUUID = u
			}
		}
	}
	if v, ok := data.ExtraData["XUID"]; ok {
		pk.ClientXUID, _ = v.(string)
	}
}

func (pk *LoginPacket) parseClientData(raw string) {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return
	}
	payload, err := decodeJWTPayload(parts[1])
	if err != nil {
		return
	}
	var data struct {
		ClientRandomID int64  `json:"ClientRandomId"`
		ServerAddress  string `json:"ServerAddress"`
		LanguageCode   string `json:"LanguageCode"`
		SkinID         string `json:"SkinId"`
		SkinData       string `json:"SkinData"`
		CapeData       string `json:"CapeData"`
		SkinGeomName   string `json:"SkinGeometryName"`
		SkinGeomData   string `json:"SkinGeometry"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return
	}
	pk.ClientID = data.ClientRandomID
	pk.ServerAddr = data.ServerAddress
	pk.Language = data.LanguageCode
	if pk.Language == "" {
		pk.Language = "en_US"
	}
	pk.SkinID = data.SkinID
	pk.GeometryName = data.SkinGeomName
	pk.SkinData, _ = base64.RawStdEncoding.DecodeString(data.SkinData)
	pk.CapeData, _ = base64.RawStdEncoding.DecodeString(data.CapeData)
	geomBytes, _ := base64.RawStdEncoding.DecodeString(data.SkinGeomData)
	pk.GeometryData = string(geomBytes)
	for len(pk.SkinData) < 8192 {
		pk.SkinData = append(pk.SkinData, 0x00)
	}
}
