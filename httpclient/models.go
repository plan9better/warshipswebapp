package httpclient

import "net/http"

type HttpClient struct {
	Client    *http.Client
	AuthToken string
}

type GameStatus struct {
	Nick           string   `json:"nick"`
	GameStatus     string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Opponent       string   `json:"opponent"`
	OpponentShots  []string `json:"opp_shots"`
	Timer          int      `json:"timer"`
	ShouldFire     bool     `json:"should_fire"`
}
type GameConfig struct {
	Wpbot  bool   `json:"wpbot"`
	Desc   string `json:"desc"`
	Nick   string `json:"nick"`
	Coords []byte `json:"coords"`
	Target string `json:"target_nick"`
}

type Desc struct {
	Desc     string `json:"desc"`
	Nick     string `json:"nick"`
	Opp_Desc string `json:"opp_desc"`
	Opponent string `json:"opponent"`
}

type Player struct {
	GameStatus string `json:"game_status"`
	Nick       string `json:"nick"`
}

type Stat struct {
	Games  int    `json:"games"`
	Nick   string `json:"nick"`
	Points int    `json:"points"`
	Rank   int    `json:"rank"`
	Wins   int    `json:"wins"`
}
