# Maze Game

A fault-tolerant peer-to-peer distributed game of treasure-collecting.

## Set up

1. Install Go
2. Install Fyne: `go install fyne.io/fyne/v2/cmd/fyne@latest`
3. Clone this repository: `git clone https://github.com/Zafei-Erin/Game.git`


## Instructions

- To start the game, you need to fire a tracker: `make tracker`.
- To join the game, simply open another tab and run: `make game id=**`. Replace ** and provide your own id.
- You can create multiple players in different tabs.
- To interact with the game, you can:
  - input 0 to refresh and 9 to exit;
  - input 1, 2, 3, 4 to go left, down, right and up.
