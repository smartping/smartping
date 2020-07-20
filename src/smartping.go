package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/cihub/seelog"
	"github.com/jakecoffman/cron"
	"github.com/smartping/smartping/src/funcs"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/http"
	"github.com/smartping/smartping/src/kv"
	//"sync"
)

// Init config
var Version = "0.8.0"

var consulEndpoint string
var version bool

func init() {
	flag.StringVar(&consulEndpoint, "consul-endpoint", "", "The consul KV store for agent self registy")
	flag.BoolVar(&version, "version", false, "show version")
	flag.Parse()
}

func main() {
	defer seelog.Flush()
	hostName, hostIP, err := kv.GetHostInfo()
	if err != nil {
		panic(err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	if consulEndpoint == "" {
		if consulEndpoint = os.Getenv("CONSUL_ENDPOINT"); consulEndpoint == "" {
			panic("cannot find consul-endpoint in flag and env")
		}
	}

	if err := kv.Registry(consulEndpoint, hostName, hostIP); err != nil {
		panic(err)
	}

	if version {
		fmt.Println(Version)
		os.Exit(0)
	}
	g.ParseConfig(Version)
	go funcs.ClearArchive()

	// start cron
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		go funcs.Ping()
		go funcs.Mapping()
		if g.Cfg.Mode["Type"] == "cloud" {
			go funcs.StartCloudMonitor()
		}
	}, "ping")
	c.AddFunc("0 0 * * * *", func() { go funcs.ClearArchive() }, "mtc")
	c.Start()

	// start http server
	go func() {
		http.StartHttp()
	}()

	// 非云模式下开启配置自动发现
	g.StartAutoDiscoveryConfig4LocalMode(consulEndpoint, Version)
	// 开启退出信号监听
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	tick := time.NewTicker(10 * time.Second)
GOOUT:
	for {
		select {
		case <-tick.C: // 定时注册agent信息
			seelog.Info("ttl registry agent info")
			if err := kv.Registry(consulEndpoint, hostName, hostIP); err != nil {
				panic(err)
			}
		case <-sigterm: // 收到退出信号，执行退出操作
			seelog.Info("terminating via signal")
			if err := kv.Delete(consulEndpoint, hostIP); err != nil {
				panic(err)
			}
			tick.Stop()
			c.Stop()
			break GOOUT
		}
	}
	seelog.Info("shutdown")
}
