
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type RequestChunkRadiusPacket struct {
	*packets.Packet
	Radius int32
}

func NewRequestChunkRadiusPacket() *RequestChunkRadiusPacket {
	return &RequestChunkRadiusPacket{Packet: packets.NewPacket(info.IDRequestChunkRadius)}
}

func (pk *RequestChunkRadiusPacket) Encode() {}

func (pk *RequestChunkRadiusPacket) Decode() {
	pk.Radius = pk.GetVarInt()
}


type ChunkRadiusUpdatedPacket struct {
	*packets.Packet
	Radius int32
}

func NewChunkRadiusUpdatedPacket(radius int32) *ChunkRadiusUpdatedPacket {
	return &ChunkRadiusUpdatedPacket{
		Packet: packets.NewPacket(info.IDChunkRadiusUpdated),
		Radius: radius,
	}
}

func (pk *ChunkRadiusUpdatedPacket) Encode() {
	pk.PutVarInt(pk.Radius)
}

func (pk *ChunkRadiusUpdatedPacket) Decode() {}
