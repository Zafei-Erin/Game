package player

import (
	"fmt"

	"github.com/Zafei-Erin/Game/internal/server"
	"github.com/Zafei-Erin/Game/types"
)

type GamePlayer struct {
	Id         string
	PlayerAddr string

	LocalGameState *types.GameState
}

func NewPlayer(id string, playerAddr string) *GamePlayer {
	return &GamePlayer{
		Id:         id,
		PlayerAddr: playerAddr,
	}
}

// listen to msg and ping from primary server
func InitServer(id string, listenAddr string, trackerAddr string) *server.Server {
	fmt.Println("init a normal server")

	server := &server.Server{
		Id:          id,
		ListenAddr:  listenAddr,
		TrackerAddr: trackerAddr,
		Msgch:       make(chan types.Message, 10),
	}
	go server.HandlePlayersConnection()
	go server.HandleMsgChan()
	return server
}
