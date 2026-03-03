package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type SetGameModePacket struct {
	*packets.Packet
	GameMode int32
}

func NewSetGameModePacket(gm int32) *SetGameModePacket {
	return &SetGameModePacket{
		Packet:   packets.NewPacket(info.IDSetGameMode),
		GameMode: gm,
	}
}

func (pk *SetGameModePacket) Encode() {
	pk.PutVarInt(pk.GameMode)
}

func (pk *SetGameModePacket) Decode() {}
