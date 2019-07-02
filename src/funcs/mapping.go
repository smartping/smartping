package funcs

import (
	"encoding/json"
	"fmt"
	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"math"
	"net"
	"strconv"
	"sync"
	"time"
)

var (
	MapLock   = new(sync.Mutex)
	MapStatus map[string][]g.MapVal
)

func Mapping() {
	var wg sync.WaitGroup
	MapStatus = map[string][]g.MapVal{}
	seelog.Debug("[func:Mapping]", g.Cfg.Chinamap)
	for tel, provDetail := range g.Cfg.Chinamap {
		for prov, _ := range provDetail {
			seelog.Debug("[func:Mapping]", g.Cfg.Chinamap[tel][prov])
			if len(g.Cfg.Chinamap[tel][prov]) > 0 {
				go MappingTask(tel, prov, g.Cfg.Chinamap[tel][prov], &wg)
				wg.Add(1)
			}
		}
	}
	wg.Wait()
	MapPingStorage()
}

//ping main function
func MappingTask(tel string, prov string, ips []string, wg *sync.WaitGroup) {
	seelog.Info("Start MappingTask " + tel + " " + prov + "..")
	statMap := []g.PingSt{}
	for _, ip := range ips {
		seelog.Debug("[func:StartChinaMapPing]", ip)
		ipaddr, err := net.ResolveIPAddr("ip", ip)
		if err == nil {
			for i := 0; i < 3; i++ {
				stat := g.PingSt{}
				stat.MinDelay = -1
				stat.LossPk = 0
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
					seelog.Debug("[func:StartChinaMapPing IcmpPing] ID:", i, " IP:", ip)
				} else {
					seelog.Debug("[func:StartChinaMapPing IcmpPing] ID:", i, " IP:", ip, " | ", err)
					stat.LossPk = stat.LossPk + 1
				}
				stat.SendPk = stat.SendPk + 1
				stat.LossPk = int((float64(stat.LossPk) / float64(stat.SendPk)) * 100)
				if stat.RevcPk > 0 {
					stat.AvgDelay = stat.AvgDelay / float64(stat.RevcPk)
				} else {
					stat.AvgDelay = 2000
				}
				statMap = append(statMap, stat)
			}
		} else {
			stat := g.PingSt{}
			stat.AvgDelay = 2000.00
			stat.MinDelay = 2000.00
			stat.MaxDelay = 2000.00
			stat.SendPk = 0
			stat.RevcPk = 0
			stat.LossPk = 100
			statMap = append(statMap, stat)
		}
	}
	fStatDetail := g.PingSt{}
	fT := 0
	effCnt := 0
	for _, stat := range statMap {
		if len(statMap) > 1 && fT < int(math.Ceil(float64(len(statMap)))/4) {
			if stat.LossPk == 3 {
				fT = fT + 1
				continue
			}
		}
		fStatDetail.MaxDelay = fStatDetail.MaxDelay + stat.MaxDelay
		fStatDetail.MinDelay = fStatDetail.MinDelay + stat.MinDelay
		fStatDetail.AvgDelay = fStatDetail.AvgDelay + stat.AvgDelay
		fStatDetail.SendPk = fStatDetail.SendPk + stat.SendPk
		fStatDetail.RevcPk = fStatDetail.RevcPk + stat.RevcPk
		fStatDetail.LossPk = fStatDetail.SendPk - fStatDetail.RevcPk
		effCnt = effCnt + 1
	}
	gMapVal := g.MapVal{}
	gMapVal.Name = tel
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", fStatDetail.AvgDelay/float64(effCnt)), 64)
	gMapVal.Value = value
	MapLock.Lock()
	MapStatus[prov] = append(MapStatus[prov], gMapVal)
	MapLock.Unlock()
	wg.Done()
	seelog.Info("Finish MappingTask " + tel + " " + prov + "..")
}

func MapPingStorage() {
	seelog.Info("Start MapPingStorage...")
	seelog.Debug(MapStatus)
	jdata, err := json.Marshal(MapStatus)
	if err != nil {
		seelog.Error("[func:StartPing] Json Error ", err)
	}
	sql := "REPLACE INTO [mappinglog] (logtime, mapjson) values('" + time.Now().Format("2006-01-02 15:04") + "','" + string(jdata) + "')"
	g.DLock.Lock()
	g.Db.Exec(sql)
	_, err = g.Db.Exec(sql)
	seelog.Debug(sql)
	if err != nil {
		seelog.Error("[func:StartPing] Sql Error ", err)
	}
	g.DLock.Unlock()
	seelog.Debug("[func:MapPingStorage] ", sql)
	seelog.Info("Finish MapPingStorage...")
}
