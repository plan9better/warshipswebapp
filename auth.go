package main

import (
	"fmt"
	"net/url"
	"warshipswebapp/httpclient"
)

func ParseConfig(form url.Values) httpclient.GameConfig {
	var cfg httpclient.GameConfig
	cfg.Nick = form.Get("name")
	cfg.Desc = form.Get("desc")
	cfg.Target = form.Get("target")
	fmt.Println(cfg)
	if form.Get("wpbot") == "on" {
		cfg.Wpbot = true
	} else {
		cfg.Wpbot = false
	}
	return cfg
}
