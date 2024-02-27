package lib

import (
	"encoding/json"
	"math/rand"

	"github.com/Zafei-Erin/Game/types"
)

func Marshal(data any) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return b
}

func GetPlayerByID(players []types.Player, playerID string) (types.Player, bool) {
	for _, player := range players {
		if player.PlayerId == playerID {
			return player, true
		}
	}
	return types.Player{}, false
}

func FindPlayerIndex(playerID string, players []types.Player) (int, bool) {
	for i, player := range players {
		if player.PlayerId == playerID {
			return i, true
		}
	}
	return -1, false
}

func GenerateRandomTreasure(gamestate *types.GameState, MazeSize int) types.Position {
	var treasure types.Position
	for {
		treasure = types.Position{X: rand.Intn(MazeSize), Y: rand.Intn(MazeSize)}
		// check conflicts
		overlap := false

		// Check for conflicts with players
		for _, playerInfo := range gamestate.Players {
			if treasure.X == playerInfo.PositionX && treasure.Y == playerInfo.PositionY {
				overlap = true
				break
			}
		}

		// Check if the position is already marked as a treasure in the Mazemap
		if gamestate.Mazemap[treasure.X][treasure.Y] == "*" {
			overlap = true
		}

		if !overlap {
			break
		}
	}
	return treasure
}
