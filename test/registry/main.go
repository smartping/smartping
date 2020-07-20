package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/kv"
)

var consulEndpoint string = "10.93.10.66:80"

func main() {
	hostName, hostIP, err := kv.GetHostInfo()
	if err != nil {
		panic(err)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	tick := time.NewTicker(10 * time.Second)
GOOUT:
	for {
		select {
		case <-tick.C:
			seelog.Info("ttl registry agent info")
			if err := kv.Registry(consulEndpoint, hostName, hostIP); err != nil {
				panic(err)
			}
		case <-sigterm:
			seelog.Info("terminating via signal")
			if err := kv.Delete(consulEndpoint, hostIP); err != nil {
				panic(err)
			}
			tick.Stop()
			break GOOUT
		}
	}
	seelog.Info("shutdown")
}
