package gui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Zafei-Erin/Game/types"
)

func CreateGameUIContent(playerId string, N int, gameState *types.GameState) fyne.CanvasObject {
	var mazeLabels []*widget.Label

	// create score labels
	scoreText := "Scores:\n"
	for _, info := range gameState.Players {
		scoreText += fmt.Sprintf("%s: %d\n", info.PlayerId, info.Score)
	}
	scoreLabel := widget.NewLabel(scoreText)

	// create servers and starttime labels
	serverInfo := widget.NewLabel(fmt.Sprintf("Primary Server: %s\nBackup Server: %s", gameState.PrimaryServer.PlayerId, gameState.BackupServer.PlayerId))
	startTimeLabel := widget.NewLabel("Start Time: " + gameState.StartTime)

	// create maze
	mazeGrid := container.NewGridWithColumns(N)

	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			cell := widget.NewLabel(" ")
			cell.Resize(fyne.NewSize(25, 25))

			mazeLabels = append(mazeLabels, cell)

			// create a border
			border := canvas.NewRectangle(color.RGBA{R: 255, G: 165, B: 0, A: 255})
			border.SetMinSize(fyne.NewSize(25, 25))

			// put border and cell in the same container
			borderedCell := container.NewStack(border, cell)

			mazeGrid.Add(borderedCell)
		}
	}

	// set player positions and treasures
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			flag := gameState.Mazemap[i][j]
			if flag == "*" {
				index := i*N + j
				mazeLabels[index].SetText("*")
			}
			if flag != "" && flag != "*" {
				index := i*N + j
				mazeLabels[index].SetText(flag)
			}
		}
	}

	// combine
	leftContainer := container.NewVBox(scoreLabel, serverInfo, startTimeLabel)
	rightContainer := mazeGrid
	content := container.NewHBox(leftContainer, rightContainer)

	return content
}
