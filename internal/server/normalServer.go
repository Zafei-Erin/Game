package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

// normal player only response when regenerate backup
func (s *Server) handleMessage(message types.Message) {
	request := types.ReqToServer{}
	if err := json.Unmarshal(message.Payload, &request); err != nil {
		fmt.Println("unmarshal error")
		return
	}

	if request.Type == "generate_backup_server" {
		gameState := types.GameState{}
		if err := json.Unmarshal(request.Data, &gameState); err != nil {
			panic(err)
		}
		gameState.BackupServer = types.PlayerAddr{
			PlayerId:   s.Id,
			PlayerAddr: s.ListenAddr,
		}
		s.gameServer = &GameServer{
			trackerAddr:    s.TrackerAddr,
			pingch:         make(chan bool),
			gamestate:      &gameState,
			msgch:          make(chan types.Message, 50),
			status:         "backup_server",
			playerChannels: make(map[string]chan types.Message),
		}
		// Start a timer to detect the aliveness of primary server.
		go s.backupReceivePing()

		// Inform the tracker of the new player list.
		msg := types.Req{
			Type: "update",
			Data: lib.Marshal(s.gameServer.gamestate.Players),
		}
		// Problem: Here we need to include trackerAddr into Servers.
		conn, err := net.Dial("tcp", s.TrackerAddr)
		if err != nil {
			fmt.Println("connect tracker failed")
			return
		}
		conn.Write(lib.Marshal(msg))
		defer conn.Close()

		fmt.Printf("receive generate_backup_server signal from %s \n", request.Id)
		message.Conn.Write([]byte("ok"))
	} else {
		message.Conn.Close()
	}
}
