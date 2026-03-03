
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

const (
	ResourcePackResponseRefused     = 1
	ResourcePackResponseSendPacks   = 2
	ResourcePackResponseAllPacksSent = 3
	ResourcePackResponseCompleted   = 4
)

type ResourcePackClientResponsePacket struct {
	*packets.Packet
	Status     byte
	PackIDs    []string
}

func NewResourcePackClientResponsePacket() *ResourcePackClientResponsePacket {
	return &ResourcePackClientResponsePacket{
		Packet: packets.NewPacket(info.IDResourcePackClientResponse),
	}
}

func (pk *ResourcePackClientResponsePacket) Encode() {}

func (pk *ResourcePackClientResponsePacket) Decode() {
	pk.Status = pk.GetByte()
	count := pk.GetLShort()
	pk.PackIDs = make([]string, count)
	for i := uint16(0); i < count; i++ {
		pk.PackIDs[i] = pk.GetString()
	}
}
