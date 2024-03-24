package server

import (
	"fmt"
	"net"
	"time"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

// only call when game start
func (s *Server) NewPrimaryServer(K int, N int) {
	fmt.Println("init primary server")

	// only primary and backup server have game server
	s.gameServer = s.NewGameServer(K, N)

	go s.gameServer.ping()
}

func (gameServer *GameServer) ping() {
	for {
		// ping every 0.5s
		time.Sleep(time.Millisecond * 500)
		alivePlayers := []types.Player{}
		deadPlayers := []types.Player{}
		alivePlayerIds := []string{}
		deadPlayerIds := []string{}
		gameServer.mu.Lock()
		// Ping all servers, get the list of alive ones.
		for index, p := range gameServer.gamestate.Players {
			// Don't ping itself.
			if index == 0 {
				alivePlayers = append(alivePlayers, p)
				alivePlayerIds = append(alivePlayerIds, p.PlayerId)
				continue
			}

			addr := p.PlayerAddr
			err := gameServer.sendPingMessage(addr)
			if err == nil {
				alivePlayers = append(alivePlayers, p)
				alivePlayerIds = append(alivePlayerIds, p.PlayerId)
			} else {
				deadPlayerIds = append(deadPlayerIds, p.PlayerId)
				deadPlayers = append(deadPlayers, p)
				fmt.Println("dead players: ", deadPlayerIds)
			}
		}

		// all alive
		if len(alivePlayers) == len(gameServer.gamestate.Players) && gameServer.gamestate.BackupServer.PlayerAddr != "" {
			gameServer.mu.Unlock()
			continue
		}

		// someone crashed
		gameServer.gamestate.Players = alivePlayers

		// Update the map, cleaning dead players.
		for _, p := range deadPlayers {
			x, y := p.PositionX, p.PositionY
			gameServer.gamestate.Mazemap[x][y] = ""
		}

		b := lib.Marshal(gameServer.gamestate)

		// update on backup server
		if gameServer.gamestate.BackupServer.PlayerAddr != "" {
			if err := gameServer.sendToBackup(b); err != nil {
				// The backup server is dead. Select and generate a new backup server.
				gameServer.gamestate.BackupServer.PlayerAddr = ""
				gameServer.gamestate.BackupServer.PlayerId = ""
			}
		}

		// old backup servre crashed, assign a new backup server
		if gameServer.gamestate.BackupServer.PlayerAddr == "" && len(alivePlayers) > 1 {
		assign_new_backup:
			for _, i := range alivePlayers[1:] {
				newBackup := types.PlayerAddr{
					PlayerId:   i.PlayerId,
					PlayerAddr: i.PlayerAddr,
				}
				if err := gameServer.assignBackupServer(newBackup); err == nil {
					gameServer.gamestate.BackupServer = newBackup
					fmt.Printf("assign %s as backup\n", i.PlayerId)
					break assign_new_backup
				}
			}
		}

		gameServer.mu.Unlock()
	}
}

func (gameServer *GameServer) sendPingMessage(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// fmt.Println("sending to player: ", err)
		return err
	}
	msg := types.ReqToServer{
		Type: "ping",
		Id:   gameServer.gamestate.PrimaryServer.PlayerId,
	}

	conn.Write(lib.Marshal(msg))
	defer conn.Close()
	return nil
}
