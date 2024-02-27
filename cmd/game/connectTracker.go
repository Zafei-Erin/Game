package main

import (
	"encoding/json"
	"io"
	"log"
	"net"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

// join game: send id and addr
func connectTracker(trackerAddr string, id string, res *types.Res) error {
	// write
	conn, err := net.Dial("tcp", trackerAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := types.Req{
		Type: "init",
		Data: []byte(id),
	}
	conn.Write(lib.Marshal(msg))

	// read
	buffer := make([]byte, 8192)
	n, readErr := conn.Read(buffer)
	if readErr != nil && readErr != io.EOF {
		log.Fatal(readErr)
	} else if readErr == io.EOF {
		return nil
	}
	if err := json.Unmarshal(buffer[:n], &res); err != nil {
		return err
	}

	return nil
}
