package main

import (
	"time"
	"warshipswebapp/httpclient"
)

func get_player_ships(httpc *httpclient.HttpClient, p *player) {
	player_ships, _ := httpc.GetGameBoard()

	p.P.Lock()
	for _, v := range player_ships {
		p.Player[v] = "ship"
	}
	p.P.Unlock()
}

func updateStatus(httpc *httpclient.HttpClient, p *player) {
	p.P.Lock()
	p.Status, _ = httpc.GetGameStatus()
	p.P.Unlock()
}
func updateBoard(p *player) {
	p.P.Lock()
	for _, v := range p.Status.OpponentShots {
		if p.Player[v] == "ship" || p.Player[v] == "hit" {
			p.Player[v] = "hit"
		} else {
			p.Player[v] = "miss"
		}
	}

	p.P.Unlock()
}

func updateEnemyBoard(e *enemy) {

	// sink entire ships not just the last part
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
}

func player_bot(httpc *httpclient.HttpClient, p *player) {
	get_player_ships(httpc, p)
	updateStatus(httpc, p)

	for p.Status.GameStatus != "ended" {
		updateStatus(httpc, p)
		for !p.Status.ShouldFire && p.Status.GameStatus != "ended" {
			time.Sleep(1 * time.Second)
			updateStatus(httpc, p)
		}

		updateBoard(p)
		time.Sleep(1 * time.Second)
	}
	updateBoard(p)
}
