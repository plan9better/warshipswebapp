package main

import (
	"fmt"
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
	return autoplay == "on"
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
