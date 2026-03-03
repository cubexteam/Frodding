
package bedrock

import (
	"github.com/cubexteam/Frodding/net/info"
	"github.com/cubexteam/Frodding/net/packets"
)

type DisconnectPacket struct {
	*packets.Packet
	HideScreen bool
	Message    string
}

func NewDisconnectPacket(message string, hideScreen bool) *DisconnectPacket {
	return &DisconnectPacket{
		Packet:     packets.NewPacket(info.IDDisconnect),
		HideScreen: hideScreen,
		Message:    message,
	}
}

func (pk *DisconnectPacket) Encode() {
	pk.PutBool(pk.HideScreen)
	if !pk.HideScreen {
		pk.PutString(pk.Message)
	}
}

func (pk *DisconnectPacket) Decode() {
	pk.HideScreen = pk.GetBool()
	if !pk.HideScreen {
		pk.Message = pk.GetString()
	}
}
