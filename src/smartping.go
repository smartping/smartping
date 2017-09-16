package main

import (
	"./funcs"
	"./g"
	"./http"
	"github.com/gy-games-libs/cron"
)

// Init config
var Version = "0.3.0"

func main() {
	config, db := g.ParseConfig(Version)
	for _, target := range config.Targets {
		go funcs.StartPing(target, db)
	}
	//go funcs.StartAlert(config, db)
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		for _, target := range config.Targets {
			go funcs.StartPing(target, db)
		}
		go funcs.StartAlert(config, db)
	}, "ping")
	c.Start()
	// HTTP
	http.StartHttp(db, &config)
}
