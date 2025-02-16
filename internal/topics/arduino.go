package topics

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
	"github.com/pocketbase/pocketbase/core"
)

type Arduino interface {
	Atmosphere(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet)
	Security(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet)
	Logging(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet)
	VoiceControl(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet)
}

type arduino struct {
	record *core.Record
}

func NewArduino(record *core.Record) Arduino {
	return &arduino{
		record: record,
	}
}

func (a *arduino) Atmosphere(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
}

func (a *arduino) Security(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
}

func (a *arduino) Logging(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
}

func (a *arduino) VoiceControl(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
}
