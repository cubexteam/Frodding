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

const (
	numSubChunks  = 8
	chunkHeight   = numSubChunks * 16
	blockBedrock  = 7
	blockStone    = 1
	blockDirt     = 3
	blockGrass    = 2
	surfaceLevel  = 63
	dirtLevel     = 62
)

var flatChunkCache []byte

func BuildFlatChunk() []byte {
	if flatChunkCache != nil {
		result := make([]byte, len(flatChunkCache))
		copy(result, flatChunkCache)
		return result
	}

	const blocksPerSub = 4096
	const nibbleSize = 2048

	buf := &bytes.Buffer{}
	buf.Grow(1 + numSubChunks*(1+blocksPerSub+nibbleSize*3) + 256 + 256 + 2)

	buf.WriteByte(numSubChunks)

	for sc := 0; sc < numSubChunks; sc++ {
		buf.WriteByte(0)
		baseY := sc * 16

		blocks := make([]byte, blocksPerSub)
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				for y := 0; y < 16; y++ {
					absY := baseY + y
					idx := x*256 + z*16 + y
					switch {
					case absY == 0:
						blocks[idx] = blockBedrock
					case absY < dirtLevel:
						blocks[idx] = blockStone
					case absY == dirtLevel:
						blocks[idx] = blockDirt
					case absY == surfaceLevel:
						blocks[idx] = blockGrass
					}
				}
			}
		}
		buf.Write(blocks)
		buf.Write(make([]byte, nibbleSize))

		skylight := make([]byte, nibbleSize)
		if baseY >= chunkHeight/2 {
			for i := range skylight {
				skylight[i] = 0xFF
			}
		}
		buf.Write(skylight)
		buf.Write(make([]byte, nibbleSize))
	}

	heightmap := bytes.Repeat([]byte{64}, 256)
	buf.Write(heightmap)

	biomes := bytes.Repeat([]byte{1}, 256)
	buf.Write(biomes)

	buf.Write([]byte{0, 0})

	flatChunkCache = buf.Bytes()
	result := make([]byte, len(flatChunkCache))
	copy(result, flatChunkCache)
	return result
}

func (pk *FullChunkDataPacket) Encode() {
	pk.PutVarInt(pk.ChunkX)
	pk.PutVarInt(pk.ChunkZ)
	pk.PutUVarInt(uint32(len(pk.Data)))
	pk.PutBytes(pk.Data)
}

func (pk *FullChunkDataPacket) Decode() {}
