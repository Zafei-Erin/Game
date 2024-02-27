package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Zafei-Erin/Game/internal/gui"
	"github.com/Zafei-Erin/Game/internal/player"
	"github.com/Zafei-Erin/Game/types"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("Wrong number of parameters...exiting")
	}

	trackerAddr := os.Args[1] + ":" + os.Args[2]
	id := os.Args[3]

	// connect with tracker
	var res types.Res
	connectTracker(trackerAddr, id, &res)

	// create player
	addr := res.Players[len(res.Players)-1].PlayerAddr
	p := player.NewPlayer(id, addr)
	initServer := player.InitServer(p.Id, p.PlayerAddr, trackerAddr)
	localGameState := types.GameState{}

	flag := false

	if len(res.Players) != 1 {
		// wait for the primary server to get ready
		time.Sleep(5 * time.Millisecond)
	}

	// connect to primary server
	for _, player := range res.Players[:len(res.Players)-1] {
		// connect with existing players
		err := p.SendToServer(player.PlayerAddr, "join", &localGameState)
		if err == nil {
			flag = true
			fmt.Printf("primary server %s approve my join request: \n", player.PlayerId)
			break
		} else {
			fmt.Println(err)
		}
	}

	// if all connection fails, transform itself to primary server
	if !flag {
		initServer.NewPrimaryServer(res.K, res.N)
		time.Sleep(2 * time.Millisecond)

		p.SendToServer(p.PlayerAddr, "join", &localGameState)
	}

	p.LocalGameState = &localGameState

	myApp := app.New()
	w := myApp.NewWindow(id)
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(500, 500))
	content := gui.CreateGameUIContent(id, res.N, p.LocalGameState)

	w.SetContent(content)

	go func() {
		for {
			var input int
			fmt.Scan(&input)
			switch input {
			case 0:
				p.SendMessageToServer("refresh", &localGameState)
			case 1:
				p.SendMessageToServer("left", &localGameState)
			case 2:
				p.SendMessageToServer("down", &localGameState)
			case 3:
				p.SendMessageToServer("right", &localGameState)
			case 4:
				p.SendMessageToServer("up", &localGameState)
			case 9:
				fmt.Printf("%s Exiting the game.", id)
				os.Exit(0)
			default:
			}
			content = gui.CreateGameUIContent(id, res.N, p.LocalGameState)
			w.SetContent(content)
		}
	}()
	w.ShowAndRun()
}
