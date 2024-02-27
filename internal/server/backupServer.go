package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

// If the backup server didn't receive ping for 550ms, it changes to the primary server.
func (s *Server) backupReceivePing() {
	for {
		isChanged := false
		timer := time.NewTimer(550 * time.Millisecond)
		select {
		case <-s.gameServer.pingch:
			fmt.Println("Primary server is alive")
			isChanged = false
		case <-timer.C:
			isChanged = true
		}
		if isChanged {
			fmt.Println("This server changes from backup to primary!")
			s.gameServer.changeToPrimary()
			break
		}
	}
}

func (gameServer *GameServer) changeToPrimary() {
	gameServer.mu.Lock()
	// change to primary
	gameServer.gamestate.PrimaryServer = gameServer.gamestate.BackupServer
	gameServer.status = "primary_server"

	// clean old primary server from the players list and gamestate
	discardedServer := gameServer.gamestate.Players[0]
	gameServer.gamestate.Players = gameServer.gamestate.Players[1:]
	gameServer.gamestate.Mazemap[discardedServer.PositionX][discardedServer.PositionY] = ""
	gameServer.gamestate.BackupServer.PlayerAddr = ""
	gameServer.gamestate.BackupServer.PlayerId = ""

	// backup server create and run player threads
	gameServer.playerChannels = make(map[string]chan types.Message)
	for _, temp := range gameServer.gamestate.Players {
		ch := make(chan types.Message, 50)
		gameServer.playerChannels[temp.PlayerId] = ch

		go gameServer.playerRoutine(ch)
	}

	// update tracker with cleaned player list
	msg := types.Req{
		Type: "update",
		Data: lib.Marshal(gameServer.gamestate.Players),
	}
	conn, err := net.Dial("tcp", gameServer.trackerAddr)
	if err != nil {
		return
	}
	conn.Write(lib.Marshal(msg))
	defer conn.Close()

	// assign backup, just choose the second in the player list
	for _, i := range gameServer.gamestate.Players[1:] {
		newBackup := types.PlayerAddr{
			PlayerId:   i.PlayerId,
			PlayerAddr: i.PlayerAddr,
		}
		if err := gameServer.assignBackupServer(newBackup); err == nil {
			fmt.Printf("Assigning %s as backup server\n", i.PlayerId)
			gameServer.gamestate.BackupServer = newBackup
			break
		} else {
			fmt.Printf("Assigning %s as backup server error: %s\n", i.PlayerId, err)
		}
	}
	gameServer.mu.Unlock()
	go gameServer.ping()
}

// create and run player threads
func (gameServer *GameServer) playerRoutine(ch chan types.Message) {
	for msg := range ch {
		request := types.ReqToServer{}
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			fmt.Println("unmarshal error: ", err)
			return
		}
		fmt.Printf("in player thread %s, move type: %s \n", request.Id, request.Type)
		gameServer.mu.Lock()

		// primary server handle player's request
		switch request.Type {
		case "refresh", "up", "down", "left", "right":
			gameServer.move(gameServer.gamestate, request.Type, request.Id)

		default:
			panic("invalid message received")
		}

		b, err := json.Marshal(gameServer.gamestate)
		if err != nil {
			panic(err)
		}

		// try to notify backend server and player at the same time
		if gameServer.gamestate.BackupServer.PlayerAddr != "" {
			go gameServer.sendToBackup(b)
		}
		msg.Conn.Write(b)

		gameServer.mu.Unlock()
	}
}

// transform a normal player to backup server
func (gameServer *GameServer) assignBackupServer(playerInfo types.PlayerAddr) error {
	maxRetries := 3

	for retry := 0; retry <= maxRetries; retry++ {
		conn, err := net.Dial("tcp", playerInfo.PlayerAddr)
		if err != nil {
			fmt.Printf("assigning backup server: error (retry %d of %d)\n", retry+1, maxRetries)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer conn.Close()

		msg := types.ReqToServer{
			Type: "generate_backup_server",
			Id:   gameServer.gamestate.PrimaryServer.PlayerId,
			Data: lib.Marshal(gameServer.gamestate),
		}

		conn.Write(lib.Marshal(msg))

		buffer := make([]byte, 8192)
		n, readErr := conn.Read(buffer)
		if readErr != nil && readErr != io.EOF {
			fmt.Println(readErr)
			return readErr
		}

		if string(buffer[:n]) != "ok" {
			return fmt.Errorf("assign backup server error")
		}
		return nil
	}

	return fmt.Errorf("assigning backup server: max retries reached, unable to connect")
}
