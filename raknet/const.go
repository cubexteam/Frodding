package raknet

import "time"

var raknetMagic = [16]byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

func RaknetMagic() []byte {
	b := raknetMagic
	return b[:]
}

const (
	MaxMTU      = 1400
	MinMTU      = 400
	MaxChannels = 32
)

const (
	SessionTimeout        = 10 * time.Second
	PingSendInterval      = 2500 * time.Millisecond
	DetectionSendInterval = 5 * time.Second
	SendInterval          = 50 * time.Millisecond
	RecoverySendInterval  = 50 * time.Millisecond
)

type Reliability byte

const (
	Unreliable                    Reliability = 0
	UnreliableSequenced           Reliability = 1
	Reliable                      Reliability = 2
	ReliableOrdered               Reliability = 3
	ReliableSequenced             Reliability = 4
	UnreliableWithACKReceipt      Reliability = 5
	ReliableWithACKReceipt        Reliability = 6
	ReliableOrderedWithACKReceipt Reliability = 7
)

func (r Reliability) IsReliable() bool {
	return r == Reliable || r == ReliableOrdered || r == ReliableSequenced ||
		r == ReliableWithACKReceipt || r == ReliableOrderedWithACKReceipt
}

func (r Reliability) IsOrdered() bool {
	return r == ReliableOrdered || r == ReliableOrderedWithACKReceipt
}

func (r Reliability) IsSequenced() bool {
	return r == UnreliableSequenced || r == ReliableSequenced
