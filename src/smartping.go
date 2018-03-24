package main

import (
	"./funcs"
	"./g"
	"./http"
	"github.com/gy-games-libs/cron"
	//"github.com/gy-games-libs/seelog"
	"flag"
	"fmt"
	"os"
	"runtime"
)

// Init config
var Version = "0.4.1"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	version := flag.Bool("v", false, "show version")
	flag.Parse()
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	config, db := g.ParseConfig(Version)
	for _, target := range config.Targets {
		go funcs.CreateDB(target, db)
	}
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		for _, target := range config.Targets {
			if target.Addr != config.Ip {
				go funcs.StartPing(target, db, config)
			}
		}
		go funcs.StartAlert(config, db)
	}, "ping")
	c.Start()
	// HTTP
	http.StartHttp(db, &config)
}
