
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type MovePlayerPacket struct {
	*packets.Packet
	RuntimeID uint64
	X, Y, Z   float32
	Pitch, Yaw, HeadYaw float32
	Mode      byte
	OnGround  bool
	RidingID  uint64
}

func NewMovePlayerPacket() *MovePlayerPacket {
	return &MovePlayerPacket{Packet: packets.NewPacket(info.IDMovePlayer)}
}

func (pk *MovePlayerPacket) Encode() {
	pk.PutUVarLong(pk.RuntimeID)
	pk.PutLFloat(pk.X)
	pk.PutLFloat(pk.Y)
	pk.PutLFloat(pk.Z)
	pk.PutLFloat(pk.Pitch)
	pk.PutLFloat(pk.Yaw)
	pk.PutLFloat(pk.HeadYaw)
	pk.PutByte(pk.Mode)
	pk.PutBool(pk.OnGround)
	pk.PutUVarLong(pk.RidingID)
}

func (pk *MovePlayerPacket) Decode() {
	pk.RuntimeID = pk.GetUVarLong()
	pk.X = pk.GetLFloat()
	pk.Y = pk.GetLFloat()
	pk.Z = pk.GetLFloat()
	pk.Pitch = pk.GetLFloat()
	pk.Yaw = pk.GetLFloat()
	pk.HeadYaw = pk.GetLFloat()
	pk.Mode = pk.GetByte()
	pk.OnGround = pk.GetBool()
	pk.RidingID = pk.GetUVarLong()
}
