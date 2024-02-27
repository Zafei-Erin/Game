package server

import (
	"fmt"
	"net"

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
