package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"warshipswebapp/httpclient"
)

var httpc *httpclient.HttpClient

type player struct {
	Player map[string]string
	P      sync.Mutex
	End    bool
	Status httpclient.GameStatus
}

type enemy struct {
	enemy map[string]string
	e     sync.Mutex
}

var e enemy

var p player

func start_game(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	resp := ""
	if r.FormValue("name") == "" {
		resp = `<h1 class="error">Error: no name provided</h1>`
		w.Write([]byte(resp))
		return
	}
	if r.FormValue("target") == "" && r.FormValue("wpbot") == "off" {
		resp = `<h1>Error: both target and wpbot are set</h1>`
		w.Write([]byte(resp))
		return
	}

	cfg := ParseConfig(r.Form)
	autoplay := ParseAutoplay(r.FormValue("autoplay"))

	token, err := httpc.GetAuthToken(&cfg)
	if err != nil {
		resp = `<h1>Error: could not get auth token</h1>`
		w.Write([]byte(resp))
		return
	}

	httpc.AuthToken = token

	// 1 - bot on bot
	// 2 - human on bot
	// 3 - human on human
	// 4 - lobby
	gameType := ChooseGameType(cfg, autoplay)
	if gameType == 0 {
		resp = `<h1>Error: could not determine game type</h1>`
		w.Write([]byte(resp))
		return
	}

	fmt.Println("redirecting")
	switch gameType {
	case 1:
		http.Redirect(w, r, "/game_player_bot", http.StatusSeeOther)
	default:
		resp = `<h1>Error: game type not bot_bot</h1>`
		w.Write([]byte(resp))
		return
	}

}

func check_turn(w http.ResponseWriter, r *http.Request) {
	res := ""
	if p.Status.GameStatus == "ended" {
		res = "<h1> Game ended </h1>"
	} else {

		if p.Status.ShouldFire {
			res = "<h1> Your turn </h1>"
		} else {
			res = "<h1> Opponent's turn </h1>"
		}
	}
	w.Write([]byte(res))
}

func lobby(w http.ResponseWriter, r *http.Request) {
	lobby := httpc.GetLobby()
	lobbyhtml := ""
	for _, v := range lobby {
		lobbyhtml += `<p class="lobby_entry">`
		lobbyhtml += v.Nick
		lobbyhtml += ":"
		lobbyhtml += v.GameStatus
		lobbyhtml += `</p>`
	}
	w.Write([]byte(lobbyhtml))
}

func fire(w http.ResponseWriter, r *http.Request) {
	field := r.Header.Get("Hx-Trigger-Name")
	if !p.Status.ShouldFire {
		return
	}

	effect, err := httpc.Fire(field)
	e.enemy[field] = effect
	if err != nil {
		w.Write([]byte("error firing"))
	}
	updateEnemyBoard(&e)
	res := `<div class="` + effect + `">` + field + `</div>`

	if effect == "miss" {
		p.Status.ShouldFire = false
	}
	w.Write([]byte(res))

}

func game_player_bot(w http.ResponseWriter, r *http.Request) {
	p.End = false
	p.Player = make(map[string]string)
	e.enemy = make(map[string]string)
	go player_bot(httpc, &p)
	http.ServeFile(w, r, "./static/board.html")
}

func enemy_board(w http.ResponseWriter, r *http.Request) {
	res := Enemy_board_to_html(e.enemy)
	w.Write([]byte(res))
}

func player_board(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(Board_to_html(p.Player)))
}

func main() {
	httpc = &httpclient.HttpClient{
		Client: &http.Client{Timeout: time.Second * 20},
	}
	http.HandleFunc("/fire", fire)
	http.HandleFunc("/check_turn", check_turn)
	http.HandleFunc("/game_player_bot", game_player_bot)
	http.HandleFunc("/start_game", start_game)
	http.HandleFunc("/lobby", lobby)
	http.HandleFunc("/player_board", player_board)
	http.HandleFunc("/enemy_board", enemy_board)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":3000", nil)
}
