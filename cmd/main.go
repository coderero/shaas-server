package main

import (
	"log"

	"coderero.dev/iot/smaas-server/internal/server"
)

func main() {
	server := server.NewServer(1883)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
