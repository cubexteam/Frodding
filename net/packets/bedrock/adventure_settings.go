
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type AdventureSettingsPacket struct {
	*packets.Packet
	Flags          uint32
	CommandPermission uint32
	Flags2         uint32
	PlayerPermission uint32
	CustomFlags    uint32
	EntityID       int64
}

func NewAdventureSettingsPacket() *AdventureSettingsPacket {
	return &AdventureSettingsPacket{Packet: packets.NewPacket(info.IDAdventureSettings)}
}

func (pk *AdventureSettingsPacket) Encode() {
	pk.PutUVarInt(pk.Flags)
	pk.PutUVarInt(pk.CommandPermission)
	pk.PutUVarInt(pk.Flags2)
	pk.PutUVarInt(pk.PlayerPermission)
	pk.PutUVarInt(pk.CustomFlags)
	pk.PutLLong(pk.EntityID)
}

func (pk *AdventureSettingsPacket) Decode() {
	pk.Flags = pk.GetUVarInt()
	pk.CommandPermission = pk.GetUVarInt()
	pk.Flags2 = pk.GetUVarInt()
	pk.PlayerPermission = pk.GetUVarInt()
	pk.CustomFlags = pk.GetUVarInt()
	pk.EntityID = pk.GetLLong()
}
