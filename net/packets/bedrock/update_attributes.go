package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type Attribute struct {
	Min     float32
	Max     float32
	Current float32
	Default float32
	Name    string
}

type UpdateAttributesPacket struct {
	*packets.Packet
	RuntimeID  uint64
	Attributes []Attribute
}

func NewUpdateAttributesPacket(runtimeID uint64) *UpdateAttributesPacket {
	return &UpdateAttributesPacket{
		Packet:    packets.NewPacket(info.IDUpdateAttributes),
		RuntimeID: runtimeID,
	}
}

func (pk *UpdateAttributesPacket) Encode() {
	pk.PutUVarLong(pk.RuntimeID)
	pk.PutUVarInt(uint32(len(pk.Attributes)))
	for _, attr := range pk.Attributes {
		pk.PutLFloat(attr.Min)
		pk.PutLFloat(attr.Max)
		pk.PutLFloat(attr.Current)
		pk.PutLFloat(attr.Default)
		pk.PutString(attr.Name)
	}
}

func (pk *UpdateAttributesPacket) Decode() {}

func DefaultAttributes() []Attribute {
	return []Attribute{
		{Min: 0, Max: 20, Current: 20, Default: 20, Name: "minecraft:health"},
		{Min: 0, Max: 20, Current: 20, Default: 20, Name: "minecraft:player.hunger"},
		{Min: 0, Max: 5, Current: 5, Default: 5, Name: "minecraft:player.saturation"},
		{Min: 0, Max: 5, Current: 0, Default: 0, Name: "minecraft:player.exhaustion"},
		{Min: 0, Max: 24791, Current: 0, Default: 0, Name: "minecraft:player.level"},
		{Min: 0, Max: 1, Current: 0, Default: 0, Name: "minecraft:player.experience"},
		{Min: 0, Max: 10, Current: 0.1, Default: 0.1, Name: "minecraft:movement"},
	}
}
