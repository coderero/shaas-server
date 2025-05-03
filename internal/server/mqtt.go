package server

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"coderero.dev/iot/smaas-server/internal/collections"
	"coderero.dev/iot/smaas-server/internal/topics"
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/pocketbase/pocketbase/core"
)

type MQTT struct {
	server  *mqtt.Server
	arduino *topics.Arduino
}

func NewMQTT(port int, collections []collections.CollectionDefiner, app core.App) *MQTT {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelError,
		})),
	})
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = server.AddHook(new(auth.Hook), &auth.Options{
		Data: config,
	})
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
		server:  server,
		arduino: topics.NewArduino(collections, app, server),
	}
}

func (m *MQTT) Start() error {
	if err := m.server.Serve(); err != nil {
		return err
	}
	return nil
}

func (m *MQTT) RegisterTopics() {
	m.arduino.RegisterTopics()
}

func loadConfig() ([]byte, error) {
	config, err := os.ReadFile("config/mqtt.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return config, nil
}
