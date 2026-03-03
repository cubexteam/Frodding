
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type ResourcePackInfoPacket struct {
	*packets.Packet
	MustAccept bool
}

func NewResourcePackInfoPacket(mustAccept bool) *ResourcePackInfoPacket {
	return &ResourcePackInfoPacket{
		Packet:     packets.NewPacket(info.IDResourcePackInfo),
		MustAccept: mustAccept,
	}
}

func (pk *ResourcePackInfoPacket) Encode() {
	pk.PutBool(pk.MustAccept)
	pk.PutLShort(0)
	pk.PutLShort(0)
}

func (pk *ResourcePackInfoPacket) Decode() {}
