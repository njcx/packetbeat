package main

import (
	"os"

	"packetbeat/libbeat/beat"
	"packetbeat/beater"

	// import protocol modules
	_ "packetbeat/include"
)

var Name = "packetbeat"

// Setups and Runs Packetbeat
func main() {
	if err := beat.Run(Name, "", beater.New); err != nil {
		os.Exit(1)
	}
}
