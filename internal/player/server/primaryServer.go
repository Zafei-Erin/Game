package server

import (
	"fmt"
)

// only call when game start
func (s *Server) NewPrimaryServer(K int, N int) {
	fmt.Println("init primary server")
	
	// only primary and backup server have game server
	s.gameServer = s.NewGameServer(K, N)
	
	go s.gameServer.ping()
}
