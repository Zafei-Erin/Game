package tracker

import "net"

type Tracker struct {
	n       int
	k       int
	players []PlayerInfo
	ln      net.Listener
	msgch   chan Message
	quitch  chan struct{}
}

type Req struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type Res struct {
	N       int          `json:"n"`
	K       int          `json:"k"`
	Players []PlayerInfo `json:"player_list"`
}

type PlayerInfo struct {
	PlayerId   string `json:"player_id"`
	PlayerAddr string `json:"player_addr"`
}

type Message struct {
	conn    net.Conn
	payload []byte
}
