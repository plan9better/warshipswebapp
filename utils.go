package main

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"warshipswebapp/httpclient"
)

func ChooseGameType(cfg httpclient.GameConfig, autoplay bool) int {
	// 1 - bot on bot
	// 2 - human on bot
	// 3 - human on human
	// 4 - lobby
	if cfg.Wpbot && autoplay {
		return 1
	}
	if cfg.Wpbot && !autoplay {
		return 2
	}
	if cfg.Target != "" {
		return 3
	}

	return 4
}

func ParseAutoplay(autoplay string) bool {
	if autoplay == "on" {
		return true
	} else {
		return false
	}
}
func Enemy_board_to_html(board map[string]string) string {
	html := ""

	for letter := 'A'; letter <= 'J'; letter++ {
		html += `<div class="row">`
		for number := 1; number <= 10; number++ {
			coord := fmt.Sprintf("%c%d", letter, number)
			if board[coord] == "" {
				html += fmt.Sprintf(`<div class="%s" hx-post="/fire" name="%s" id="%s">%s</div>`, board[coord], coord, coord, coord)
			} else {
				html += fmt.Sprintf(`<div class="%s" id="%s">%s</div>`, board[coord], coord, coord)
			}
		}
		html += `</div>`
	}
	return html
}

func Board_to_html(board map[string]string) string {
	html := ""

	for letter := 'A'; letter <= 'J'; letter++ {
		html += `<div class="row">`
		for number := 1; number <= 10; number++ {
			coord := fmt.Sprintf("%c%d", letter, number)
			html += fmt.Sprintf(`<div class="%s" id="%s">%s</div>`, board[coord], coord, coord)
		}
		html += `</div>`
	}
	return html
}

// func Stats_to_html(stats []ttpclient.Stat) string{
// 	for stat := range stats{

// 	}
// }

func strToCoord(str string) Coord {
	c := []byte(str)
	var coord Coord
	coord.X = int(c[0])
	if len(c) == 3 {
		coord.Y = 10
	} else {
		coord.Y = int(str[1] - '0')
	}
	return coord
}

type Coord struct {
	X int
	Y int
}

func (c Coord) toString() string {
	return fmt.Sprintf("%d%d", c.X, c.Y)
}

func (c Coord) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d,%d", c.X, c.Y)), nil
}

func (c *Coord) UnmarshalText(text []byte) error {
	textStr := string(text)

	txt := strings.Split(textStr, ",")

	x, err := strconv.Atoi(txt[0])
	if err != nil {
		panic(err)
	}

	y, err := strconv.Atoi(txt[1])
	if err != nil {
		panic(err)
	}
	c.X = x
	c.Y = y

	return nil
}

type Ship []Coord

type Player struct {
	Ships          []Ship
	LastMove       Coord
	LastMoveEffect int
	HasLastMove    bool
}

// const (
// 	hit  = 0
// 	miss = 1
// 	sunk = 2
// )

func IsSame(cord1 Coord, cord2 Coord) bool {
	if cord1.X == cord1.Y && cord2.X == cord2.Y {
		return true
	}
	return false
}

func IsAdjacent(cord1 Coord, cord2 Coord) bool {
	if IsSame(cord1, cord2) {
		return false
	}

	if math.Abs(float64(cord1.X-cord2.X)) <= 1 && math.Abs(float64(cord1.Y-cord2.Y)) <= 1 {
		return true
	}
	return false
}

func whichEdge(coord Coord) string {
	if coord.X == 'A' {
		return "l"
	}
	if coord.X == 'J' {
		return "r"
	}
	if coord.Y == 10 {
		return "t"
	}
	if coord.Y == 0 {
		return "b"
	}
	return ""
}

func whichCorner(coord Coord) string {
	// bottom left
	if coord.X == 'A' && coord.Y == 1 {
		return "bl"
	}
	// top left
	if coord.X == 'A' && coord.Y == 10 {
		return "tl"
	}
	// top right
	if coord.X == 'J' && coord.Y == 10 {
		return "tr"
	}
	// bottom right
	if coord.X == 'J' && coord.Y == 1 {
		return "br"
	}
	return ""
}

func isInCorner(coord Coord) bool {
	// bottom left
	if coord.X == 'A' && coord.Y == 1 {
		return true
	}
	// top left
	if coord.X == 'A' && coord.Y == 10 {
		return true
	}
	// top right
	if coord.X == 'J' && coord.Y == 10 {
		return true
	}
	// bottom right
	if coord.X == 'J' && coord.Y == 1 {
		return true
	}
	return false
}

func isOnEdge(coord Coord) bool {
	if coord.Y == 1 || coord.Y == 10 || coord.X == 'A' || coord.Y == 'J' {
		return true
	}
	return false
}

func FindToShoot(coord Coord) []Coord {
	var res []Coord
	if isInCorner(coord) {
		switch whichCorner(coord) {
		case "tl":
			res = append(res, Coord{X: 'A', Y: 9})
			res = append(res, Coord{X: 'B', Y: 10})
			return res
		case "tr":
			res = append(res, Coord{X: 'J', Y: 9})
			res = append(res, Coord{X: 'I', Y: 10})
			return res
		case "bl":
			res = append(res, Coord{X: 'A', Y: 2})
			res = append(res, Coord{X: 'B', Y: 1})
			return res
		case "br":
			res = append(res, Coord{X: 'J', Y: 2})
			res = append(res, Coord{X: 'I', Y: 1})
			return res

		default:
			log.Println("func FindAdjacent if isInCorner hit default case in switch?", whichCorner(coord))
		}
	}

	if isOnEdge(coord) {
		switch whichEdge(coord) {
		case "t":
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			return res
		case "b":
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			return res
		case "l":
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			return res
		case "r":
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			return res
		}
	}

	res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
	res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
	res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
	res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
	return res
}

func FindAdjacent(coord Coord) []Coord {

	var res []Coord

	if isInCorner(coord) {
		switch whichCorner(coord) {
		case "tl":
			res = append(res, Coord{X: 'A', Y: 9})
			res = append(res, Coord{X: 'B', Y: 9})
			res = append(res, Coord{X: 'B', Y: 10})
			return res
		case "tr":
			res = append(res, Coord{X: 'J', Y: 9})
			res = append(res, Coord{X: 'I', Y: 9})
			res = append(res, Coord{X: 'I', Y: 10})
			return res
		case "bl":
			res = append(res, Coord{X: 'A', Y: 2})
			res = append(res, Coord{X: 'B', Y: 2})
			res = append(res, Coord{X: 'B', Y: 1})
			return res
		case "br":
			res = append(res, Coord{X: 'J', Y: 2})
			res = append(res, Coord{X: 'I', Y: 2})
			res = append(res, Coord{X: 'I', Y: 1})
			return res

		default:
			log.Println("func FindAdjacent if isInCorner hit default case in switch?", whichCorner(coord))
		}
	}

	if isOnEdge(coord) {
		switch whichEdge(coord) {
		case "t":
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			return res
		case "b":
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			return res
		case "l":
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X + 1, Y: coord.Y - 1})
			return res
		case "r":
			res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y + 1})
			res = append(res, Coord{X: coord.X - 1, Y: coord.Y - 1})
			return res
		}
	}

	res = append(res, Coord{X: coord.X + 1, Y: coord.Y + 1})
	res = append(res, Coord{X: coord.X + 1, Y: coord.Y})
	res = append(res, Coord{X: coord.X + 1, Y: coord.Y - 1})
	res = append(res, Coord{X: coord.X, Y: coord.Y + 1})
	res = append(res, Coord{X: coord.X, Y: coord.Y - 1})
	res = append(res, Coord{X: coord.X - 1, Y: coord.Y + 1})
	res = append(res, Coord{X: coord.X - 1, Y: coord.Y})
	res = append(res, Coord{X: coord.X - 1, Y: coord.Y - 1})
	return res
}

func (p *Player) LogShot(shot Coord, shotEffect int) {
	p.HasLastMove = true
	p.LastMove = shot
	p.LastMoveEffect = shotEffect
}
