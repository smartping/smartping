package funcs

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"net"
	"strconv"
	"sync"
	"time"
)

func Ping() {
	var wg sync.WaitGroup
	for _, target := range g.Cfg.Targets {
		if target.Addr != g.Cfg.Ip {
			wg.Add(1)
			go PingTask(target, &wg)
		}
	}
	wg.Wait()
	go StartAlert()
}

//ping main function
func PingTask(t g.Target, wg *sync.WaitGroup) {
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
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, "IP:", t.Addr, "| ", err)
			} else {
				seelog.Debug("[func:StartPing IcmpPing] ID:", i, "IP:", t.Addr, "| ", err)
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
	PingStorage(stat, t)
	wg.Done()
	seelog.Info("Finish Ping " + t.Addr + "..")
}

//storage ping data
func PingStorage(pingres g.PingSt, t g.Target) {
	checktime := time.Now().Format("2006-01-02 15:04")
	l := g.PingLog{}
	l.Logtime = checktime
	l.Maxdelay = strconv.FormatFloat(pingres.MaxDelay, 'f', 2, 64)
	l.Mindelay = strconv.FormatFloat(pingres.MinDelay, 'f', 2, 64)
	l.Avgdelay = strconv.FormatFloat(pingres.AvgDelay, 'f', 2, 64)
	l.Losspk = strconv.Itoa(pingres.LossPk)
	seelog.Info("[func:StartPing] ", "(", checktime, ")Starting runPingTest ", t.Name)
	db := g.GetDb("ping", t.Addr)
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("pinglog"))
		if err != nil {
			return fmt.Errorf("create bucket error : %s", err)
		}
		jdata, _ := json.Marshal(l)
		err = b.Put([]byte(checktime[8:]), []byte(string(jdata)))
		if err != nil {
			return fmt.Errorf("put data error: %s", err)
		}
		return nil
	})
	if err != nil {
		seelog.Error("[func:StoragePing] Data Storage Error: ", err)
	}
	seelog.Info("[func:StartPing] ", "(", checktime, ") PingTest on ", t.Name, " finish!")
}
