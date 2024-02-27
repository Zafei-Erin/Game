package tracker

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/Zafei-Erin/Game/types"
)

type Tracker struct {
	n       int
	k       int
	players []types.PlayerAddr
	ln      net.Listener
	msgch   chan types.Message
	quitch  chan struct{}
}

func NewTracker(n, k int) *Tracker {
	return &Tracker{
		n:       n,
		k:       k,
		players: []types.PlayerAddr{},
		msgch:   make(chan types.Message, 50),
	}
}

func (t *Tracker) Listen(port string) error {
	ln, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Println("tracker listening", port)
	t.ln = ln

	go t.accept()

	<-t.quitch

	return nil
}

func (t *Tracker) accept() {
	for {
		conn, acceptErr := t.ln.Accept()
		if acceptErr != nil {
			fmt.Println("accept error: ", acceptErr)
			continue
		}
		defer conn.Close()

		buf := make([]byte, 8192)
		n, readErr := conn.Read(buf)
		if readErr != nil {
			fmt.Println("read error: ", readErr)
			return
		}

		t.msgch <- types.Message{
			Conn:    conn,
			Payload: buf[:n],
		}
	}
}

func (t *Tracker) HandleMsgChan() {
	for msg := range t.msgch {
		t.handleMsg(msg)
	}
}

func (t *Tracker) handleMsg(message types.Message) {
	request := types.Req{}
	if err := json.Unmarshal(message.Payload, &request); err != nil {
		fmt.Println("unmarshal error")
		return
	}

	switch request.Type {
	case "init":
		fmt.Printf("%s tries to join game", request.Data)
		t.players = append(t.players, types.PlayerAddr{
			PlayerId:   string(request.Data),
			PlayerAddr: message.Conn.RemoteAddr().String(),
		})

		res := types.Res{
			N:       t.n,
			K:       t.k,
			Players: t.players,
		}

		b, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		message.Conn.Write(b)

	case "update":
		fmt.Printf("These players are cleaned: %s \n", t.players)
		json.Unmarshal(request.Data, &t.players)
	default:
	}
}
