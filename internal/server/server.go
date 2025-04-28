package server

import (
	"golang.org/x/sync/errgroup"
)

type Server struct {
	mqttServer       *MQTT
	pocketbaseServer *PocketBase
}

func NewServer(mqttPort int) *Server {
	pocketbaseServer := NewPocketBase()
	return &Server{
		pocketbaseServer: pocketbaseServer,
		mqttServer:       NewMQTT(mqttPort, pocketbaseServer.GetCollectionsNames(), pocketbaseServer.app),
	}
}

func (s *Server) Start() error {
	s.pocketbaseServer.RegisterRoutes()
	s.pocketbaseServer.RegisterMigrations()
	s.mqttServer.RegisterTopics()

	var g errgroup.Group

	g.Go(func() error {
		if err := s.pocketbaseServer.Start(); err != nil {
			return err
		}

		return nil
	})

	// Run the MQTT server in a goroutine
	g.Go(func() error {
		if err := s.mqttServer.Start(); err != nil {
			return err
		}

		return nil
	})

	// Wait for both servers to finish or return an error if any fails
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
