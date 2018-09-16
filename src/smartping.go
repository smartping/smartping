package main

import (
	"flag"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/funcs"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/http"
	"github.com/jakecoffman/cron"
	"os"
	"runtime"
	"sync"
)

// Init config
var Version = "0.6.0"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	version := flag.Bool("v", false, "show version")
	flag.Parse()
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	g.ParseConfig(Version)

	for _, target := range g.Cfg.Targets {
		go funcs.CreatePingTable(target)
	}
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		var wg sync.WaitGroup
		for _, target := range g.Cfg.Targets {
			if target.Addr != g.Cfg.Ip {
				wg.Add(1)
				go funcs.StartPing(target, &wg)
			}
		}
		wg.Wait()
		go funcs.StartAlert()
		seelog.Info(g.Cfg.Mode)
		if g.Cfg.Mode == "cloud" {
			go funcs.StartCloudMonitor(1)
		}
	}, "ping")
	c.AddFunc("0 0 0 * * *", func() {
		go funcs.ClearAlertTable()
		go funcs.ClearPingTable()
	}, "mtc")
	c.Start()
	http.StartHttp()
}
