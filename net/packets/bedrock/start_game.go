package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type StartGamePacket struct {
	*packets.Packet
	EntityUniqueID       int64
	EntityRuntimeID      uint64
	PlayerGameMode       int32
	SpawnX               float32
	SpawnY               float32
	SpawnZ               float32
	Yaw                  float32
	Pitch                float32
	Seed                 int32
	Dimension            int32
	Generator            int32
	GameMode             int32
	Difficulty           int32
	SpawnBX              int32
	SpawnBY              int32
	SpawnBZ              int32
	AchievementsDisabled bool
	Time                 int32
	EduMode              bool
	RainLevel            float32
	LightningLevel       float32
	IsMultiplayer        bool
	BroadcastToLAN       bool
	CommandsEnabled      bool
	ForceTexturePacks    bool
	LevelName            string
	IsTrial              bool
	CurrentTick          int64
	EnchantmentSeed      int32
}

func NewStartGamePacket() *StartGamePacket {
	return &StartGamePacket{
		Packet: packets.NewPacket(info.IDStartGame),
	}
}

func (pk *StartGamePacket) Encode() {
	pk.PutVarLong(pk.EntityUniqueID)
	pk.PutUVarLong(pk.EntityRuntimeID)
	pk.PutVarInt(pk.PlayerGameMode)
	pk.PutLFloat(pk.SpawnX)
	pk.PutLFloat(pk.SpawnY)
	pk.PutLFloat(pk.SpawnZ)
	pk.PutLFloat(pk.Yaw)
	pk.PutLFloat(pk.Pitch)
	pk.PutVarInt(pk.Seed)
	pk.PutVarInt(pk.Dimension)
	pk.PutVarInt(pk.Generator)
	pk.PutVarInt(pk.GameMode)
	pk.PutVarInt(pk.Difficulty)
	pk.PutVarInt(pk.SpawnBX)
	pk.PutUVarInt(uint32(pk.SpawnBY))
	pk.PutVarInt(pk.SpawnBZ)
	pk.PutBool(pk.AchievementsDisabled)
	pk.PutVarInt(pk.Time)
	pk.PutBool(pk.EduMode)
	pk.PutLFloat(pk.RainLevel)
	pk.PutLFloat(pk.LightningLevel)
	pk.PutBool(pk.IsMultiplayer)
	pk.PutBool(pk.BroadcastToLAN)
	pk.PutBool(false)
	pk.PutBool(pk.CommandsEnabled)
	pk.PutBool(pk.ForceTexturePacks)
	pk.PutUVarInt(0)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutVarInt(1)
	pk.PutLInt(4)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutBool(false)
	pk.PutString(pk.LevelName)
	pk.PutString("")
	pk.PutBool(pk.IsTrial)
	pk.PutLLong(pk.CurrentTick)
	pk.PutVarInt(pk.EnchantmentSeed)
}

func (pk *StartGamePacket) Decode() {}
