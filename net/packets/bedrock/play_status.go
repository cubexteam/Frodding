
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

const (
	PlayStatusLoginSuccess    int32 = 0
	PlayStatusLoginFailedClient int32 = 1
	PlayStatusLoginFailedServer int32 = 2
	PlayStatusPlayerSpawn     int32 = 3
	PlayStatusLoginFailedInvalidTenant int32 = 4
	PlayStatusLoginFailedVanillaEdu    int32 = 5
	PlayStatusLoginFailedEduVanilla    int32 = 6
)

type PlayStatusPacket struct {
	*packets.Packet
	Status int32
}

func NewPlayStatusPacket(status int32) *PlayStatusPacket {
	pk := &PlayStatusPacket{
		Packet: packets.NewPacket(info.IDPlayStatus),
		Status: status,
	}
	return pk
}

func (pk *PlayStatusPacket) Encode() {
	pk.PutInt(pk.Status)
}

func (pk *PlayStatusPacket) Decode() {
	pk.Status = pk.GetInt()
}
