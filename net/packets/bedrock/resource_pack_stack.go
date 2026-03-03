
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type ResourcePackStackPacket struct {
	*packets.Packet
	MustAccept bool
}

func NewResourcePackStackPacket(mustAccept bool) *ResourcePackStackPacket {
	return &ResourcePackStackPacket{
		Packet:     packets.NewPacket(info.IDResourcePackStack),
		MustAccept: mustAccept,
	}
}

func (pk *ResourcePackStackPacket) Encode() {
	pk.PutBool(pk.MustAccept)
	pk.PutLShort(0)
	pk.PutLShort(0)
}

func (pk *ResourcePackStackPacket) Decode() {}
