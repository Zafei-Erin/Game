package server 

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

// If the backup server didn't receive ping for 2000ms, it changes to the primary server.
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

	// Remove the former primary from tracker
	// Also discard it from the players list and the gamestate
	discardedServer := gameServer.gamestate.Players[0]
	gameServer.gamestate.Players = gameServer.gamestate.Players[1:]
	gameServer.gamestate.Mazemap[discardedServer.PositionX][discardedServer.PositionY] = ""
	gameServer.gamestate.BackupServer.PlayerAddr = ""
	gameServer.gamestate.BackupServer.PlayerId = ""

	gameServer.playerChannels = make(map[string]chan types.Message)
	for _, temp := range gameServer.gamestate.Players {
		ch := make(chan types.Message, 50)
		gameServer.playerChannels[temp.PlayerId] = ch

		go gameServer.playerRoutine(ch)
	}

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
	for indexi, i := range gameServer.gamestate.Players {
		if indexi == 0 {
			continue
		}
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

func (gameServer *GameServer) playerRoutine(ch chan types.Message) {

	for msg := range ch {
		// fmt.Printf("new msg!")
		request := types.ReqToServer{}
		if err := json.Unmarshal(msg.Payload, &request); err != nil {
			fmt.Println("unmarshal error: ", err)
			return
			// panic(err)
		}
		fmt.Printf("in player thread %s, move type: %s \n", request.Id, request.Type)
		gameServer.mu.Lock()
		defer gameServer.mu.Unlock()

		// primary server
		switch request.Type {
		case "refresh", "up", "down", "left", "right":
			gameServer.move(gameServer.gamestate, request.Type, request.Id)

		default:
			panic("invalid message received")
		}

		//gameServer.gamestate.Test += 1

		b, err := json.Marshal(gameServer.gamestate)
		if err != nil {
			panic(err)
		}

		if gameServer.gamestate.BackupServer.PlayerAddr != "" {
			go gameServer.sendToBackup(b)
		}

		gameServer.mu.Unlock()
		msg.Conn.Write(b)

	}
}
