package funcs

import (
	"../g"
	"../ping"
	"github.com/gy-games-libs/seelog"
	"net"
	"sync"
	"time"
)

func StartPing(t g.Target, wg *sync.WaitGroup) {
	seelog.Info("Start Ping " + t.Addr + "..")
	stat := g.PingSt{}
	var ip, _ = net.ResolveIPAddr("ip", t.Addr)
	if ip == nil {
		seelog.Error("[func:IcmpPing] Finish Addr:", ip, " Domain or Ip not valid!")
		wg.Done()
		return
	}
	stat.MinDelay = -1
	lossPK := 0
	for i := 0; i < 20; i++ {
		starttime := time.Now().UnixNano()
		delay, err := ping.SendICMP(ip, i)
		if err == nil {
			stat.AvgDelay = stat.AvgDelay + delay
			if stat.MaxDelay < delay {
				stat.MaxDelay = delay
			}
			if stat.MinDelay == -1 || stat.MinDelay > delay {
				stat.MinDelay = delay
			}
			stat.RevcPk = stat.RevcPk + 1
		} else {
			seelog.Debug("[func:IcmpPing] ID:", i, " | ", err)
			lossPK = lossPK + 1
		}
		stat.SendPk = stat.SendPk + 1
		stat.LossPk = int((float64(lossPK) / float64(stat.SendPk)) * 100)
		duringtime := time.Now().UnixNano() - starttime
		time.Sleep(time.Duration(3000*1000000-duringtime) * time.Nanosecond)
	}
	stat.AvgDelay = stat.AvgDelay / float64(stat.SendPk)
	seelog.Debug("[func:IcmpPing] Finish Addr:", ip, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	StoragePing(stat, t)
	wg.Done()
	seelog.Info("Finish Ping " + t.Addr + "..")
}
