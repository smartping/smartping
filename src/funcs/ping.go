package funcs

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
)

func Ping() {
	var wg sync.WaitGroup
	seelog.Infof("Ping===>>targets: %v", g.SelfCfg.Ping)
	for _, target := range g.SelfCfg.Ping {
		seelog.Infof("Ping===>>target: %s", target)
		wg.Add(1)
		go PingTask(g.Cfg.Network[target], &wg)
	}
	wg.Wait()
	go StartAlert()
}

//ping main function
func PingTask(t g.NetworkMember, wg *sync.WaitGroup) {
	seelog.Info("Start Ping " + t.Addr + "..")
	stat := g.PingSt{}
	stat.MinDelay = -1
	lossPK := 0
	ipaddr, err := net.ResolveIPAddr("ip", t.Addr)
	if err == nil {
		for i := 0; i < 20; i++ {
			starttime := time.Now().UnixNano()
			delay, err := nettools.RunPing(ipaddr, 3*time.Second, 64, i)
			if err == nil {
				stat.AvgDelay = stat.AvgDelay + delay
				if stat.MaxDelay < delay {
					stat.MaxDelay = delay
				}
				if stat.MinDelay == -1 || stat.MinDelay > delay {
					stat.MinDelay = delay
				}
				stat.RevcPk = stat.RevcPk + 1
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, " IP:", t.Addr)
			} else {
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, " IP:", t.Addr, "| err:", err)
				lossPK = lossPK + 1
			}
			stat.SendPk = stat.SendPk + 1
			stat.LossPk = int((float64(lossPK) / float64(stat.SendPk)) * 100)
			duringtime := time.Now().UnixNano() - starttime
			time.Sleep(time.Duration(3000*1000000-duringtime) * time.Nanosecond)
		}
		if stat.RevcPk > 0 {
			stat.AvgDelay = stat.AvgDelay / float64(stat.RevcPk)
		} else {
			stat.AvgDelay = 0.0
		}
		seelog.Debug("[func:IcmpPing] Finish Addr:", t.Addr, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	} else {
		stat.AvgDelay = 0.00
		stat.MinDelay = 0.00
		stat.MaxDelay = 0.00
		stat.SendPk = 0
		stat.RevcPk = 0
		stat.LossPk = 100
		seelog.Debug("[func:IcmpPing] Finish Addr:", t.Addr, " Unable to resolve destination host")
	}
	PingStorage(stat, t.Addr)
	wg.Done()
	seelog.Info("Finish Ping " + t.Addr + "..")
}

//storage ping data
func PingStorage(pingres g.PingSt, Addr string) {
	logtime := time.Now().Format("2006-01-02 15:04")
	seelog.Info("[func:StartPing] ", "(", logtime, ")Starting PingStorage ", Addr)
	sql := "INSERT INTO [pinglog] (logtime, target, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk) values('" + logtime + "','" + Addr + "','" + strconv.FormatFloat(pingres.MaxDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.MinDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.AvgDelay, 'f', 2, 64) + "','" + strconv.Itoa(pingres.SendPk) + "','" + strconv.Itoa(pingres.RevcPk) + "','" + strconv.Itoa(pingres.LossPk) + "')"
	seelog.Debug("[func:StartPing] ", sql)
	g.DLock.Lock()
	_, err := g.Db.Exec(sql)
	if err != nil {
		seelog.Error("[func:StartPing] Sql Error ", err)
	}
	g.DLock.Unlock()
	seelog.Info("[func:StartPing] ", "(", logtime, ") Finish PingStorage  ", Addr)
}
