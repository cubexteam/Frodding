package bedrock

import (
	"bytes"

	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type FullChunkDataPacket struct {
	*packets.Packet
	ChunkX int32
	ChunkZ int32
	Data   []byte
}

func NewFullChunkDataPacket(cx, cz int32, data []byte) *FullChunkDataPacket {
	return &FullChunkDataPacket{
		Packet: packets.NewPacket(info.IDFullChunkData),
		ChunkX: cx,
		ChunkZ: cz,
		Data:   data,
	}
}

func BuildFlatChunk() []byte {
	buf := &bytes.Buffer{}

	const numSubChunks = 8
	buf.WriteByte(numSubChunks)

	for sc := 0; sc < numSubChunks; sc++ {
		buf.WriteByte(0)

		blocks := make([]byte, 4096)
		baseY := sc * 16

		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				for y := 0; y < 16; y++ {
					absY := baseY + y
					idx := x*256 + z*16 + y
					if absY == 0 {
						blocks[idx] = 7
					} else if absY < 62 {
						blocks[idx] = 1
					} else if absY == 62 {
						blocks[idx] = 3
					} else if absY == 63 {
						blocks[idx] = 2
					}
				}
			}
		}
		buf.Write(blocks)

		buf.Write(make([]byte, 2048))

		skylight := make([]byte, 2048)
		for i := range skylight {
			if baseY >= 64 {
				skylight[i] = 0xFF
			} else {
				skylight[i] = 0x00
			}
		}
		buf.Write(skylight)

		buf.Write(make([]byte, 2048))
	}

	heightmap := make([]byte, 256)
	for i := range heightmap {
		heightmap[i] = 64
	}
	buf.Write(heightmap)

	biomes := make([]byte, 256)
	for i := range biomes {
		biomes[i] = 1
	}
	buf.Write(biomes)

	buf.WriteByte(0)
	buf.WriteByte(0)

	return buf.Bytes()
}

func (pk *FullChunkDataPacket) Encode() {
	pk.PutVarInt(pk.ChunkX)
	pk.PutVarInt(pk.ChunkZ)
	pk.PutUVarInt(uint32(len(pk.Data)))
	pk.PutBytes(pk.Data)
}

func (pk *FullChunkDataPacket) Decode() {}
