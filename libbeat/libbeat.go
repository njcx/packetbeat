package main

import (
	"os"

	"packetbeat/libbeat/beat"
	"packetbeat/libbeat/mock"
)

func main() {
	if err := beat.Run(mock.Name, mock.Version, mock.New); err != nil {
		os.Exit(1)
	}
}
