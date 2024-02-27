package server

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

type Server struct {
	Id          string
	ListenAddr  string
	TrackerAddr string
	Msgch       chan types.Message

	gameServer *GameServer
	quitch     chan struct{}
	ln         net.Listener
}

func (s *Server) HandlePlayersConnection() error {
	fmt.Println("listen to: " + s.ListenAddr)
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		fmt.Println("server listen error ")
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.readMsgFromPlayer()

	<-s.quitch
	return nil
}

// accept connection from players
func (s *Server) readMsgFromPlayer() {
	for {
		conn, acceptErr := s.ln.Accept()
		if acceptErr != nil {
			// do not return, keep accepting new conn
			fmt.Println("accept error: ", acceptErr)
			continue
		}
		defer conn.Close()

		buf := make([]byte, 8192)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			fmt.Println("read error: ", readErr)
			continue
		}

		s.Msgch <- types.Message{
			Conn:    conn,
			Payload: buf[:n],
		}

	}
}

func (s *Server) HandleMsgChan() {
	for msg := range s.Msgch {
		// for normal player
		if s.gameServer == nil {
			s.handleMessage(msg)
			continue
		} else {
			// for primary server and backup server
			s.gameServer.handleMessage(msg)
			continue
		}
	}
}

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
