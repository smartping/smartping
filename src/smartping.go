package main

import (
	"./funcs"
	"./g"
	"./http"
	"github.com/gy-games-libs/cron"
	"github.com/gy-games-libs/seelog"
	"os"
	"runtime"
	"flag"
	"fmt"
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
		if config.Ping == "sysping" {
			for _, target := range config.Targets {
				if target.Addr !=config.Ip{
					go funcs.StartSysPing(target, db, config)
				}
			}
		} else if config.Ping == "goping" {
			for _, target := range config.Targets {
				if target.Addr !=config.Ip {
					go funcs.StartGoPing(target, db, config)
				}
			}
		} else {
			seelog.Error("[Init] Ping Method Error!")
			os.Exit(0)
		}
		go funcs.StartAlert(config, db)
	}, "ping")
	c.Start()
	// HTTP
	http.StartHttp(db, &config)
}
