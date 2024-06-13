package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const MAX_TRIES = 5
const TIMEOUT = 5

func (c *HttpClient) makeRequest(endpoint string, v any, method string, payload io.Reader) error {
	address := fmt.Sprintf("https://go-pjatk-server.fly.dev/api/%s", endpoint)
	req, err := http.NewRequest(method, address, payload)
	if err != nil {
		log.Printf("Error making a get request to %s\n", endpoint)
		log.Printf("Error: %s\n", err)
		return err
	}

	req.Header.Add("X-Auth-Token", c.AuthToken)

	var body []byte

	isCritical := false
	for tries := 1; (tries <= MAX_TRIES) && (!isCritical); tries++ {
		resp, err := c.Client.Do(req)
		if err != nil {
			log.Printf("Error sending a get request to %s\n", endpoint)
			log.Printf("Error: %s\n", err)
			continue
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body from request to %s\n", endpoint)
			log.Printf("Error: %s\n", err)
			continue
		}

		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, body, endpoint)
			continue
		}

		err = json.Unmarshal(body, &v)
		if err != nil {
			log.Printf("Error unmarshaling JSON response: %s\n", err)
			continue
		}
		return nil

	}

	log.Println("Failed after ", MAX_TRIES, " tries., ", endpoint)
	return err

}

func handleHTTPCodes(code int, body []byte, endpoint string) bool {
	log.Printf("_______________________________________\n")
	log.Printf("HTTP Code %d while requesting %s\n", code, endpoint)
	if string(body) == "" {
		log.Println("(Response body is empty)")
	} else {
		log.Println(string(body))
	}

	isCritical := false
	switch code {
	case 400:
		log.Printf("Bad request\n")

		// Check if the bad request is because not your turn
		// if so wait 1 second and then try again
		if string(body) == `{"message":"not your turn"}` {
			isCritical = false
			time.Sleep(time.Second)
		} else {
			isCritical = true
		}

	case 404:
		log.Printf("Not found")
		isCritical = true
	case 401:
		log.Printf("No auth")
		isCritical = true

	// Commented out since 503 is supposed
	// to simulate an unexpected error
	//
	// case 503:
	// 	log.Println("tries: ", tries)
	// 	isCritical = false
	case 429:
		// too many requests
		time.Sleep(TIMEOUT * time.Second)
		isCritical = false
	default:
		log.Println("Unexpexted error, retrying...")
		isCritical = false
	}
	return isCritical
}

func (c *HttpClient) GetGameStatus() (GameStatus, error) {
	var status GameStatus
	err := c.makeRequest("game", &status, "GET", nil)

	for err != nil {
		log.Println("Error getting game status: ", err)
		return status, err
	}

	return status, nil
}

func (c *HttpClient) GetDesc() (Desc, error) {
	var desc Desc

	for desc.Opp_Desc == "" {
		time.Sleep(time.Second)
		err := c.makeRequest("game/desc", &desc, "GET", nil)
		if err != nil {
			log.Println("Error getting description: ", err)
			return desc, err
		}
	}
	return desc, nil
}

func (c *HttpClient) GetLobby() []Player {
	var players []Player
	err := c.makeRequest("lobby", &players, "GET", nil)
	if err != nil {
		log.Println("Error requesting lobby")
	}
	return players
}

func (c *HttpClient) GetAuthToken(cfg *GameConfig) (string, error) {
	var resp *http.Response

	isCritical := false
	for tries := 1; (tries <= MAX_TRIES) && (!isCritical); tries++ {
		bm, err := json.Marshal(cfg)
		if err != nil {
			log.Println("Error marshaling request for auth token", err)
			continue
		}

		req, err := http.NewRequest("POST", "https://go-pjatk-server.fly.dev/api/game", bytes.NewReader(bm))
		if err != nil {
			log.Println("Error creating a request", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err = c.Client.Do(req)
		if err != nil {
			log.Println("Error sending request to auth")
			continue
		}

		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, nil, "auth")
			continue
		}
		return resp.Header.Get("X-Auth-Token"), nil
	}

	log.Fatal("Error creating auth token, the game cannot continue")
	return "", nil

}

func (c *HttpClient) RefreshWaitSession() {
	req, err := http.NewRequest("GET", "https://go-pjatk-server.fly.dev/api/game/refresh", nil)
	if err != nil {
		log.Println("Error creating a request to refresh", err)
	}
	req.Header.Add("X-Auth-Token", c.AuthToken)

	var resp *http.Response
	isCritical := false
	for tries := 1; (tries <= MAX_TRIES) && (!isCritical); tries++ {
		resp, err = c.Client.Do(req)
		if err != nil {
			log.Printf("Error sending request to refresh\n")
		}
		if resp.StatusCode != 200 {
			isCritical = handleHTTPCodes(resp.StatusCode, nil, "auth")
		} else {
			break
		}
	}
}

func (c *HttpClient) Stats() []Stat {
	type stats struct {
		Stats []Stat `json:"stats"`
	}
	var s stats
	err := c.makeRequest("stats", &s, "GET", nil)
	if err != nil {
		log.Println("Error checking stats")
	}

	return s.Stats
}

func (c *HttpClient) GetGameBoard() ([]string, error) {
	type board struct {
		Board []string `json:"board"`
	}
	var brd board
	err := c.makeRequest("game/board", &brd, "GET", nil)
	if err != nil {
		log.Println("Error fetching game board")
		return brd.Board, err
	}
	return brd.Board, nil
}

func (c *HttpClient) Fire(toFire string) (string, error) {
	type coord struct {
		Coord string `json:"coord"`
	}
	var crd coord
	crd.Coord = toFire

	crdm, err := json.Marshal(crd)
	if err != nil {
		log.Println("Error marshaling fire coords")
	}

	type result struct {
		Result string `json:"result"`
	}
	var res result
	err = c.makeRequest("game/fire", &res, "POST", bytes.NewReader(crdm))
	if err != nil {
		log.Printf("Error firing: %s\n", err)
		return res.Result, err
	}
	return res.Result, nil
}

func (c *HttpClient) Abandon() {
	req, _ := http.NewRequest("DELETE", "https://go-pjatk-server.fly.dev/api/abandon", nil)
	req.Header.Add("X-Auth-Token", c.AuthToken)
	c.Client.Do(req)

}
