package player

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Zafei-Erin/Game/lib"
	"github.com/Zafei-Erin/Game/types"
)

func (player *GamePlayer) SendToServer(serverAddr string, Type string, response *types.GameState) error {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	var data any
	if Type == "join" {
		data = types.PlayerAddr{
			PlayerAddr: player.PlayerAddr,
			PlayerId:   player.Id,
		}
	}

	msg := types.ReqToServer{
		Type: Type,
		Id:   player.Id,
		Data: lib.Marshal(data),
	}

	conn.Write(lib.Marshal(msg))

	buffer := make([]byte, 8192)
	n, readErr := conn.Read(buffer)
	if readErr != nil {
		return readErr
	}
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return err
	}

	return nil
}

func (player *GamePlayer) SendMessageToServer(Type string, response *types.GameState) {
	var wg sync.WaitGroup
	fmt.Printf("send request %s to primary %s\n", Type, player.LocalGameState.PrimaryServer.PlayerId)
	// send to primary
	if err := player.SendToServer(player.LocalGameState.PrimaryServer.PlayerAddr, Type, response); err != nil {
		fmt.Printf("send request %s to primary failed, send to every one\n", Type)
		fmt.Println(response.Players)
		time.Sleep(550 * time.Millisecond)
		// if primary failed, send to everyone, only backup response
		for _, p := range response.Players {
			addr := p.PlayerAddr
			wg.Add(1)
			go func(serverAddr string, Type string, response *types.GameState, wg *sync.WaitGroup) {
				defer wg.Done()
				err := player.SendToServer(addr, Type, response)
				if err == nil {
					fmt.Printf("%s response my request %s, %d \n", p.PlayerId, Type)
				} else {
					fmt.Printf("%s reject my request %s, %d, error: %s\n", p.PlayerId, Type, err)
				}
			}(addr, Type, response, &wg)
		}
		wg.Wait()
	}
}
