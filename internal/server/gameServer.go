package server

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

type GameServer struct {
	gamestate *types.GameState

	trackerAddr    string
	mu             sync.RWMutex
	msgch          chan types.Message
	pingch         chan bool
	status         string
	playerChannels map[string]chan types.Message
}

func (s *Server) NewGameServer(K int, N int) *GameServer {
	gameState := &types.GameState{
		Players:   []types.Player{},
		StartTime: time.Now().Format("15:04:05"),
		PrimaryServer: types.PlayerAddr{
			PlayerId:   s.Id,
			PlayerAddr: s.ListenAddr,
		},
		BackupServer: types.PlayerAddr{
			PlayerId:   "",
			PlayerAddr: "",
		},
	}

	mazemap := make([][]string, N)
	for i := range mazemap {
		mazemap[i] = make([]string, N)
	}

	placedStars := 0
	for placedStars < K {
		x := rand.Intn(N)
		y := rand.Intn(N)

		if mazemap[x][y] == "" {
			mazemap[x][y] = "*"
			placedStars++
		}
	}

	gameState.Mazemap = mazemap

	gameServer := &GameServer{
		gamestate:      gameState,
		mu:             sync.RWMutex{},
		msgch:          make(chan types.Message, 50),
		pingch:         make(chan bool),
		status:         "primary_server",
		playerChannels: make(map[string]chan types.Message),
	}
	return gameServer
}

func (gameServer *GameServer) sendToBackup(marshaledGameState []byte) error {
	conn, err := net.Dial("tcp", gameServer.gamestate.BackupServer.PlayerAddr)
	if err != nil {
		fmt.Printf("sending to backup: %s \n, error: %s", gameServer.gamestate.BackupServer.PlayerId, err)
		return err
	}
	defer conn.Close()

	msg := types.ReqToServer{
		Type: "backup",
		Id:   gameServer.gamestate.PrimaryServer.PlayerId,
		Data: marshaledGameState,
	}

	conn.Write(lib.Marshal(msg))
	return nil
}

func (gameServer *GameServer) handleMessage(message types.Message) {
	request := types.ReqToServer{}
	if err := json.Unmarshal(message.Payload, &request); err != nil {
		fmt.Println("unmarshal error: ", err)
		return
	}

	// backup server only backup
	if gameServer.status == "backup_server" && request.Type == "backup" {
		gamestate := types.GameState{}
		if err := json.Unmarshal(request.Data, &gamestate); err != nil {
			panic(err)
		}
		gameServer.gamestate = &gamestate
		return
	}

	if gameServer.status == "backup_server" && request.Type == "ping" {
		gameServer.pingch <- true
		return
	}

	if gameServer.status == "backup_server" && request.Type == "join" {
		// check primary
		if err := gameServer.checkonPrimary(); err == nil {
			// if alive, return
			message.Conn.Close()
			return
		}
		// primary is failed, stop pingch, change to primary, and do not return
		close(gameServer.pingch)
		gameServer.changeToPrimary()
	}

	// backup server does not response to over type of request
	if gameServer.status == "backup_server" {
		message.Conn.Close()
		return
	}

	if gameServer.status == "primary_server" {
		fmt.Printf("receive msg from %s, move: %s\n", request.Id, request.Type)
		switch request.Type {
		case "join":
			playerInfo := types.PlayerAddr{}
			if err := json.Unmarshal(request.Data, &playerInfo); err != nil {
				panic(err)
			}
			gameServer.addPlayer(playerInfo)

			if gameServer.gamestate.BackupServer.PlayerAddr == "" && len(gameServer.gamestate.Players) == 2 {
				// backup if its the second player
				// this section would never be reached again, ping would handle
				gameServer.mu.Lock()
				if err := gameServer.assignBackupServer(playerInfo); err == nil {
					fmt.Printf("Assigning %s as backup server\n", playerInfo.PlayerId)
					gameServer.gamestate.BackupServer = playerInfo
				} else {
					fmt.Printf("Assigning %s as backup server error: %s\n", playerInfo.PlayerId, err)
				}
				gameServer.mu.Unlock()
			}

			b, err := json.Marshal(gameServer.gamestate)
			if err != nil {
				panic(err)
			}

			// send to player and backup at the same time
			if gameServer.gamestate.BackupServer.PlayerAddr != "" {
				go gameServer.sendToBackup(b)
			}
			message.Conn.Write(b)

		case "refresh", "up", "down", "left", "right":
			ch := gameServer.playerChannels[request.Id]
			ch <- message
			return

		case "ping":
			return

		default:
			fmt.Println("Invalid msg:" + string(message.Payload))
			return
		}
	}
}

func (gameServer *GameServer) checkonPrimary() error {
	addr := gameServer.gamestate.PrimaryServer.PlayerAddr
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("checkon primary error: ", err)
		return err
	}
	defer conn.Close()
	return nil
}

func (gameServer *GameServer) addPlayer(player types.PlayerAddr) {
	gameState := gameServer.gamestate
	N := len(gameState.Mazemap)

	for {
		x := rand.Intn(N)
		y := rand.Intn(N)

		if gameState.Mazemap[x][y] == "" {
			gameState.Mazemap[x][y] = player.PlayerId

			gameState.Players = append(gameState.Players, types.Player{
				PlayerId:   player.PlayerId,
				PlayerAddr: player.PlayerAddr,
				Score:      0,
				PositionX:  x,
				PositionY:  y,
			})
			break
		}
	}

	ch := make(chan types.Message, 50)
	gameServer.playerChannels[player.PlayerId] = ch
	fmt.Printf("join in %s \n", player.PlayerId)
	go gameServer.playerRoutine(ch)
}

func (gameServer *GameServer) move(gamestate *types.GameState, direction string, playerID string) {
	// Find the index of the player in the slice
	playerIndex, found := lib.FindPlayerIndex(playerID, gamestate.Players)
	if !found {
		fmt.Printf("PlayerIndex %s not found.\n", playerID)
		return
	}

	// get player position
	playerInfo, exists := lib.GetPlayerByID(gamestate.Players, playerID)
	if !exists {
		fmt.Printf("PlayerID %s not found.\n", playerID)
		return
	}

	MazeSize := len(gamestate.Mazemap)

	// update via direction
	newX, newY := playerInfo.PositionX, playerInfo.PositionY
	oldX, oldY := newX, newY

	switch direction {
	case "left":
		newY--
	case "up":
		newX--
	case "right":
		newY++
	case "down":
		newX++
	case "refresh":
	}

	flag := false

	// check bounds
	if newX < 0 || newX >= MazeSize || newY < 0 || newY >= MazeSize {
		fmt.Printf("Player %s tried to move out of bounds.\n", playerID)
		flag = true
		newX, newY = oldX, oldY
	}

	// check player conflicts
	for _, info := range gamestate.Players {
		if info.PositionX == newX && info.PositionY == newY && info.Score >= 0 && info.PlayerId != playerID {
			fmt.Printf("Player %s tried to move into a cell already occupied by another player.\n", playerID)
			flag = true
			newX, newY = oldX, oldY
		}
	}

	if gamestate.Mazemap[newX][newY] == "*" {
		fmt.Printf("Player %s collected a treasure!\n", playerID)
		playerInfo.Score++
		// Remove the collected treasure by updating the Mazemap
		gamestate.Mazemap[newX][newY] = playerID

		// generate new treasure (update Mazemap to indicate a new treasure)
		newTreasure := lib.GenerateRandomTreasure(gamestate, MazeSize)
		gamestate.Mazemap[newTreasure.X][newTreasure.Y] = "*"
	}

	if !flag {
		gamestate.Mazemap[oldX][oldY] = ""
		gamestate.Mazemap[newX][newY] = playerID

		// update player info
		playerInfo.PositionX = newX
		playerInfo.PositionY = newY
		gamestate.Players[playerIndex] = playerInfo
	}
}
