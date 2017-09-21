package funcs

import (
	"fmt"
	"github.com/gy-games-libs/go-fastping"
	"github.com/gy-games-libs/seelog"
	"net"
	"os"
	"strconv"
	"time"
)

type PingSt struct {
	SendPk   string
	RevcPk   string
	LossPk   string
	MinDelay string
	AvgDelay string
	MaxDelay string
}

func Ping(Addr string) PingSt {
	var rt PingSt
	var allcost time.Duration
	var minDelay time.Duration
	var maxDelay time.Duration
	revc := 0
	allcost = 0
	minDelay = time.Duration(3000) * time.Millisecond
	maxDelay = 0
	for i := 0; i < 20; i++ {
		p := fastping.NewPinger()
		ra, err := net.ResolveIPAddr("ip4:icmp", Addr)
		if err == nil {
			p.MaxRTT = time.Second * 3
			p.AddIPAddr(ra)
			p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
				if minDelay > rtt {
					minDelay = rtt
				}
				if maxDelay < rtt {
					maxDelay = rtt
				}
				allcost = allcost + rtt
				revc = revc + 1
			}
			err = p.Run()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	rt.MaxDelay = strconv.FormatFloat((float64(maxDelay.Nanoseconds()) / float64(1000000)), 'f', 2, 64)
	if minDelay == 3000 {
		minDelay = 0
	}
	rt.MinDelay = strconv.FormatFloat((float64(minDelay.Nanoseconds()) / float64(1000000)), 'f', 2, 64)
	if revc > 0 {
		rt.AvgDelay = strconv.FormatFloat(((float64(allcost.Nanoseconds()) / 1000000) / float64(revc)), 'f', 2, 64)
	} else {
		rt.AvgDelay = "0"
	}
	rt.RevcPk = strconv.Itoa(revc)
	rt.SendPk = "20"
	rt.LossPk = strconv.Itoa(((20 - revc) / 20) * 100)
	seelog.Debug("[func:Ping] ", "MaxDelay:"+rt.MaxDelay+" MinDelay:"+rt.MinDelay+" AvgDelay:"+rt.AvgDelay+" SendPK:"+rt.SendPk+" RevcPk:"+rt.RevcPk+" LossPK:"+rt.LossPk)
	return rt
}
