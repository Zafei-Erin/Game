package types

import "net"

type Res struct {
	N       int          `json:"n"`
	K       int          `json:"k"`
	Players []PlayerAddr `json:"player_list"`
}

type PlayerAddr struct {
	PlayerId   string `json:"player_id"`
	PlayerAddr string `json:"player_addr"`
}

type Req struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

type Message struct {
	Conn    net.Conn
	Payload []byte
}

type Position struct {
	X int
	Y int
}

type ReqToServer struct {
	Type string `json:"type"`
	Id   string `json:"id"`
	Data []byte `json:"data"`
}

type GameState struct {
	Players       []Player   `json:"all_players"`
	StartTime     string     `json:"start_time"`
	Mazemap       [][]string `json:"mazemap"`
	PrimaryServer PlayerAddr `json:"primary_server"`
	BackupServer  PlayerAddr `json:"backup_server"`
}

type Player struct {
	PlayerId   string `json:"player_id"`
	PlayerAddr string `json:"player_addr"`
	Score      int    `json:"player_score"`
	PositionX  int    `json:"position_x"`
	PositionY  int    `json:"position_y"`
}
