package main

import (
	"./funcs"
	"./g"
	"./http"
	"github.com/gy-games-libs/cron"
	"runtime"
//	"fmt"
	"github.com/gy-games-libs/seelog"
	"os"
)

// Init config
var Version = "0.4.0"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	config, db := g.ParseConfig(Version)
	for _, target := range config.Targets {
		go funcs.CreateDB(target, db)
	}
	//go funcs.StartAlert(config, db)
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		if config.Ping =="sysping"{
			for _, target := range config.Targets {
				go funcs.StartSysPing(target, db,config)
			}
		}else if config.Ping =="goping"{
			for _, target := range config.Targets {
				go funcs.StartGoPing(target, db,config)
			}
		}else if config.Ping !="fping"{
			go funcs.StartFPing(db,config)
		}else{
			seelog.Error("[Init] Ping Method Error!")
			os.Exit(0)
		}
		go funcs.StartAlert(config, db)
	}, "ping")
	c.Start()
	// HTTP
	http.StartHttp(db, &config)
}
