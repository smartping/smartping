package funcs

import (
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"sync"
	"time"
	"encoding/json"
	"github.com/boltdb/bolt"
	"fmt"
	"strconv"
)

//ping main function
func StartPing(t g.Target, wg *sync.WaitGroup) {
	seelog.Info("Start Ping " + t.Addr + "..")
	stat := g.PingSt{}
	stat.MinDelay = -1
	lossPK := 0
	for i := 0; i < 20; i++ {
		starttime := time.Now().UnixNano()
		delay, err := nettools.RunPing(t.Addr, 3*time.Second, 64, i)
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
	if stat.RevcPk > 0 {
		stat.AvgDelay = stat.AvgDelay / float64(stat.RevcPk)
	} else {
		stat.AvgDelay = 0.0
	}
	seelog.Debug("[func:IcmpPing] Finish Addr:", t.Addr, " MaxDelay:", stat.MaxDelay, " MinDelay:", stat.MinDelay, " AvgDelay:", stat.AvgDelay, " Revc:", stat.RevcPk, " LossPK:", stat.LossPk)
	StoragePing(stat, t)
	wg.Done()
	seelog.Info("Finish Ping " + t.Addr + "..")
}

//storage ping data
func StoragePing(pingres g.PingSt, t g.Target) {
	checktime := time.Now().Format("2006-01-02 15:04")
	l :=g.PingLog{}
	l.Logtime = checktime
	l.Maxdelay = strconv.FormatFloat(pingres.MaxDelay, 'f', 2, 64)
	l.Mindelay = strconv.FormatFloat(pingres.MinDelay, 'f', 2, 64)
	l.Avgdelay = strconv.FormatFloat(pingres.AvgDelay, 'f', 2, 64)
	l.Losspk = strconv.Itoa(pingres.LossPk)
	seelog.Info("[func:StartPing] ", "(", checktime, ")Starting runPingTest ", t.Name)
	db:=g.GetDb("ping",t.Addr)
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("pinglog"))
		if err != nil {
			return fmt.Errorf("create bucket error : %s", err)
		}
		jdata,_ :=json.Marshal(l)
		err = b.Put([]byte(checktime[8:]), []byte(string(jdata)))
		if err != nil {
			return fmt.Errorf("put data error: %s", err)
		}
		return nil
	})
	if err != nil {
		seelog.Error("[func:StoragePing] Data Storage Error: ",err)
	}
	seelog.Info("[func:StartPing] ", "(", checktime, ") PingTest on ", t.Name, " finish!")
}