
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type SetTimePacket struct {
	*packets.Packet
	Time int32
}

func NewSetTimePacket(time int32) *SetTimePacket {
	return &SetTimePacket{
		Packet: packets.NewPacket(info.IDSetTime),
		Time:   time,
	}
}

func (pk *SetTimePacket) Encode() {
	pk.PutVarInt(pk.Time)
}

func (pk *SetTimePacket) Decode() {}
