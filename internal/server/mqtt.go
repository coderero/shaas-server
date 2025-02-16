package server

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type MQTT struct {
	server *mqtt.Server
}

func NewMQTT(port int) *MQTT {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
	})

	err := server.AddHook(new(auth.AllowHook), nil)
	if err != nil {
		log.Fatal(err)
	}

	err = server.AddListener(listeners.NewTCP(listeners.Config{
		Type:    "tcp",
		ID:      "tcp",
		Address: fmt.Sprintf("0.0.0.0:%d", port),
	}))

	if err != nil {
		log.Fatal(err)
	}

	return &MQTT{
		server: server,
	}
}

func (m *MQTT) Start() error {
	errChan := make(chan error)
	go func() {
		err := m.server.Serve()
		if err != nil {
			errChan <- err
		}
	}()

	return <-errChan
}

func (m *MQTT) RegisterTopics() {
	m.server.Subscribe("arduino/+/+", 0, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		log.Printf("Received message on topic %s: %s", sub.ShareName[0], pk.Payload)
	})
}
