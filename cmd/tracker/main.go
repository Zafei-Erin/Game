package main

import (
	"log"
	"os"
	"strconv"

	"github.com/Zafei-Erin/Game/internal/tracker"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("Wrong number of parameters...exiting")
	}

	port := os.Args[1]
	n, _ := strconv.Atoi(os.Args[2])
	k, _ := strconv.Atoi(os.Args[3])

	tracker := tracker.NewTracker(n, k)

	go tracker.HandleMsgChan()
	tracker.Listen(port)
}
