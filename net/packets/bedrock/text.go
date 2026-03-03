
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

const (
	TextTypeRaw       = 0
	TextTypeChat      = 1
	TextTypeTranslation = 2
	TextTypePopup     = 3
	TextTypeTip       = 4
	TextTypeSystem    = 5
	TextTypeWhisper   = 6
	TextTypeAnnouncement = 7
)

type TextPacket struct {
	*packets.Packet
	TextType byte
	Source   string
	Message  string
	XUID     string
}

func NewTextPacket(textType byte, source, message string) *TextPacket {
	return &TextPacket{
		Packet:   packets.NewPacket(info.IDText),
		TextType: textType,
		Source:   source,
		Message:  message,
	}
}

func (pk *TextPacket) Encode() {
	pk.PutByte(pk.TextType)
	switch pk.TextType {
	case TextTypeChat, TextTypeWhisper, TextTypeAnnouncement:
		pk.PutString(pk.Source)
		pk.PutString(pk.Message)
	case TextTypeRaw, TextTypeTip, TextTypeSystem, TextTypePopup:
		pk.PutString(pk.Message)
	case TextTypeTranslation:
		pk.PutString(pk.Message)
		pk.PutVarInt(0)
	}
	pk.PutString(pk.XUID)
}

func (pk *TextPacket) Decode() {
	pk.TextType = pk.GetByte()
	switch pk.TextType {
	case TextTypeChat, TextTypeWhisper, TextTypeAnnouncement:
		pk.Source = pk.GetString()
		pk.Message = pk.GetString()
	default:
		pk.Message = pk.GetString()
	}
}
