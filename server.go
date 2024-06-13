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
	Player  map[string]string
	P       sync.Mutex
	End     bool
	Status  httpclient.GameStatus
	ShotSum int
	HitSum  int
	Time    int
}

var desc httpclient.Desc

type enemy struct {
	enemy map[string]string
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

	e.enemy = make(map[string]string)

	fmt.Println("redirecting")
	switch gameType {
	case 1:
		http.Redirect(w, r, "/game_bot", http.StatusSeeOther)
	case 2:
		http.Redirect(w, r, "/game_player", http.StatusSeeOther)
	case 3:
		http.Redirect(w, r, "/game_player", http.StatusSeeOther)
	case 4:
		http.Redirect(w, r, "/inlobby", http.StatusSeeOther)

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

func top10(w http.ResponseWriter, r *http.Request) {
	stats := httpc.Stats()
	statshtml := ""
	statshtml += "<table class=\"stats\">"
	statshtml += "<tr><td>Nick</td><td>Games</td><td>Points</td><td>Rank</td><td>Wins</td></tr>"
	for _, v := range stats {
		statshtml += "</tr>"
		statshtml += fmt.Sprintf("<td>%s</td>", v.Nick)
		statshtml += fmt.Sprintf("<td>%d</td>", v.Games)
		statshtml += fmt.Sprintf("<td>%d</td>", v.Points)
		statshtml += fmt.Sprintf("<td>%d</td>", v.Rank)
		statshtml += fmt.Sprintf("<td>%d</td>", v.Wins)
		statshtml += "</tr>"
	}
	w.Write([]byte(statshtml))
}
func lobby(w http.ResponseWriter, r *http.Request) {
	lobby := httpc.GetLobby()
	lobbyhtml := ""
	for _, v := range lobby {
		lobbyhtml += fmt.Sprintf("<p class=\"lobby_entry\">%s: %s</p>", v.Nick, v.GameStatus)
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
	for crd, effect := range e.enemy {
		coord := strToCoord(crd)
		tries := 0
		for effect == "sunk" && tries < 5 {
			adj := FindAdjacent(coord)
			for _, val := range adj {
				strval := val.toString()
				if e.enemy[strval] == "hit" || e.enemy[strval] == "sunk" {
					e.enemy[strval] = "sunk"
					coord = val
				} else {
					e.enemy[strval] = "miss"
				}
			}
			tries++
		}
	}

	if effect == "miss" {
		p.Status.ShouldFire = false
	}
	res := fmt.Sprintf("<div class=\"%s\">%s</div>", effect, field)

	p.ShotSum += 1
	if effect == "sunk" || effect == "hit" {
		p.HitSum += 1
	}
	p.Time = 60

	w.Write([]byte(res))

}
func shotStats(w http.ResponseWriter, r *http.Request) {
	var percent float64
	if p.ShotSum > 0 {
		percent = (float64(p.HitSum) / float64(p.ShotSum)) * 100
	} else {
		percent = 0
	}
	res := fmt.Sprintf("<p id=\"shotStat\">%.2f%%</p>", percent)
	w.Write([]byte(res))
}

func timeLeft(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%d", p.Time)))
}
func game_player(w http.ResponseWriter, r *http.Request) {
	desc, _ = httpc.GetDesc()
	p.Time = 60
	p.End = false
	go func() {
		for p.Status.GameStatus != "ended" {
			p.Time -= 1
			time.Sleep(1 * time.Second)
		}
	}()
	p.Player = make(map[string]string)
	e.enemy = make(map[string]string)
	go player_bot(httpc, &p)
	http.ServeFile(w, r, "./static/board.html")
}

func enemy_board(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(Enemy_board_to_html(e.enemy)))
}

func player_board(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(Board_to_html(p.Player)))
}
func abandon(w http.ResponseWriter, r *http.Request) {
	httpc.Abandon()
	w.Write([]byte("Game abandoned"))
	p.End = true
}

func enemyInfo(w http.ResponseWriter, r *http.Request) {
	res := fmt.Sprintf("<h1>%s</h1>", desc.Opponent)
	res += fmt.Sprintf("<h2>%s</h2>", desc.Opp_Desc)
	w.Write([]byte(res))
}
func playerInfo(w http.ResponseWriter, r *http.Request) {
	res := fmt.Sprintf("<h1>%s</h1>", desc.Nick)
	res += fmt.Sprintf("<h2>%s</h2>", desc.Desc)
	w.Write([]byte(res))
}

func main() {
	httpc = &httpclient.HttpClient{
		Client: &http.Client{Timeout: time.Second * 20},
	}
	http.HandleFunc("/timeLeft", timeLeft)
	http.HandleFunc("/fire", fire)
	http.HandleFunc("/enemyInfo", enemyInfo)
	http.HandleFunc("/playerInfo", playerInfo)
	http.HandleFunc("/abandon", abandon)
	http.HandleFunc("/check_turn", check_turn)
	http.HandleFunc("/game_player", game_player)
	http.HandleFunc("/start_game", start_game)
	http.HandleFunc("/lobby", lobby)
	http.HandleFunc("/stats", top10)
	http.HandleFunc("/shotStats", shotStats)
	http.HandleFunc("/player_board", player_board)
	http.HandleFunc("/enemy_board", enemy_board)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(":3000", nil)
}
