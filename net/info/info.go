package info

const (
	LatestProtocol       = 113
	MinProtocol          = 105
	LatestGameVersion    = "v1.1.7"
	LatestGameVersionNetwork = "1.1.7"
)

var SupportedProtocols = map[int32]string{
	105: "1.1.0",
	107: "1.1.3",
	110: "1.1.5",
	111: "1.1.6",
	113: "1.1.7",
}

func IsSupported(protocol int32) bool {
	_, ok := SupportedProtocols[protocol]
	return ok
}

const (
	IDLogin                      = 0x01
	IDPlayStatus                 = 0x02
	IDServerHandshake            = 0x03
	IDClientHandshake            = 0x04
	IDDisconnect                 = 0x05
	IDResourcePackInfo           = 0x06
	IDResourcePackStack          = 0x07
	IDResourcePackClientResponse = 0x08
	IDText                       = 0x09
	IDSetTime                    = 0x0a
	IDStartGame                  = 0x0b
	IDAddPlayer                  = 0x0c
	IDAddEntity                  = 0x0d
	IDRemoveEntity               = 0x0e
	IDAddItemEntity              = 0x0f
	IDTakeItemEntity             = 0x11
	IDMoveEntity                 = 0x12
	IDMovePlayer                 = 0x13
	IDPlayerAction               = 0x24
	IDSetEntityData              = 0x27
	IDSetEntityMotion            = 0x28
	IDUpdateAttributes           = 0x1d
	IDInventoryTransaction       = 0x1e
	IDMobEquipment               = 0x1f
	IDMobArmorEquipment          = 0x20
	IDInteract                   = 0x21
	IDBlockPickRequest           = 0x22
	IDEntityPickRequest          = 0x23
	IDAnimate                    = 0x2c
	IDRespawn                    = 0x2d
	IDContainerOpen              = 0x2e
	IDContainerClose             = 0x2f
	IDPlayerHotbar               = 0x30
	IDInventoryContent           = 0x31
	IDInventorySlot              = 0x32
	IDContainerSetData           = 0x33
	IDCraftingData               = 0x34
	IDCraftingEvent              = 0x35
	IDAdventureSettings          = 0x37
	IDBlockEntityData            = 0x38
	IDFullChunkData      = 0x3a
	IDSetGameMode        = 0x3b
	IDSetCommandsEnabled         = 0x3b
	IDSetDifficulty              = 0x3c
	IDChangeDimension            = 0x3d
	IDSetPlayerGameType          = 0x3e
	IDPlayerList                 = 0x3f
	IDRequestChunkRadius         = 0x45
	IDChunkRadiusUpdated         = 0x46
	IDAvailableCommands          = 0x4c
	IDCommandRequest             = 0x4d
	IDResourcePackDataInfo       = 0x52
	IDResourcePackChunkData      = 0x53
	IDResourcePackChunkRequest   = 0x54
	IDTransfer                   = 0x55
	IDSetTitle                   = 0x58
	IDPlayerSkin                 = 0x5d
)
