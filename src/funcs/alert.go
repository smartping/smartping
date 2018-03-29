package funcs

import (
	"../g"
	"../nettools"
	"bytes"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"strconv"
	"time"
)

//alert main function
func StartAlert() {
	seelog.Info("[func:StartAlert] ", "starting run AlertCheck ")
	timeStartStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
	dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
	sql := `CREATE TABLE IF NOT EXISTS [alertlog-` + dateStartStr + `] (
			logtime   VARCHAR (8),
			fromname  VARCHAR (15),
			toname    VARCHAR (15),
			tracert	  TEXT
	);`
	for _, v := range g.Cfg.Targets {
		if v.Addr != g.Cfg.Ip {
			checktimeStartStr := time.Unix((time.Now().Unix() - int64(v.Thdchecksec)), 0).Format("2006-01-02 15:04")
			g.DLock.Lock()
			querysql := "SELECT ifnull(max(avgdelay),0) maxavgdelay, ifnull(max(losspk),0) maxlosspk ,count(1) Cnt FROM  `pinglog-" + v.Addr + "` where lastcheck > '" + checktimeStartStr + "' and (cast(avgdelay as double) > " + strconv.Itoa(v.Thdavgdelay) + " or cast(losspk as double) > " + strconv.Itoa(v.Thdloss) + ") "
			rows, err := g.Db.Query(querysql)
			g.DLock.Unlock()
			seelog.Debug("[func:StartAlert] ", querysql)
			if err != nil {
				seelog.Error("[func:StartAlert] Query Error ", err)
				continue
			}
			for rows.Next() {
				l := new(g.TopoLog)
				err := rows.Scan(&l.Maxavgdelay, &l.Maxlosspk, &l.Cnt)
				if err != nil {
					seelog.Error("[func:StartAlert] Rows Error ", err)
					continue
				}
				sec, _ := strconv.Atoi(l.Cnt)
				if sec >= v.Thdoccnum {
					tracrtString := ""
					hops, err := nettools.RunTrace(v.Addr, time.Second, 60, 3)
					if nil != err {
						seelog.Error("[func:StartAlert] Traceroute error ", err)
						tracrtString = err.Error()
					} else {
						tracrt := bytes.NewBufferString("")
						for i, hop := range hops {
							if hop.Addr == nil {
								fmt.Fprintf(tracrt, "* * *\n")
							} else {
								fmt.Fprintf(tracrt, "%d (%s) %v %v %v\n", i+1, hop.Addr, hop.MaxRTT, hop.AvgRTT, hop.MinRTT)
							}
						}
						tracrtString = tracrt.String()
					}
					sql = sql + "insert into [alertlog-" + dateStartStr + "] (logtime,fromname,toname,tracert) values('" + timeStartStr + "','" + g.Cfg.Name + "','" + v.Name + "','" + tracrtString + "');"
				}
			}
			rows.Close()
		}
	}
	SqlExec(sql)
	seelog.Info("[func:StartAlert] ", "AlertCheck finish ")
}
