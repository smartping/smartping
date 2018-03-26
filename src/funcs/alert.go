package funcs

import (
	"../g"
	"github.com/gy-games-libs/seelog"
	"strconv"
	"time"
)

func ClearDb() {
	seelog.Info("[func:clearDb] ", "starting run clearDb ")
	sql := ""
	reminList := map[string]bool{}
	for i := 0; i < g.Cfg.Alerthistory; i++ {
		reminList["alertlog-"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")] = true
	}
	listpreout := []string{}
	lrows, lerr := g.Db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'")
	if lerr == nil {
		for lrows.Next() {
			var l string
			err := lrows.Scan(&l)
			if err != nil {
				seelog.Error("[StartAlert] ", err)
			}
			listpreout = append(listpreout, l)
		}
		lrows.Close()
	}
	for _, v := range listpreout {
		if _, ok := reminList[v]; !ok {
			sql = sql + "DROP TABLE [" + v + "];"
		}
	}
	SqlExec(sql)
	seelog.Info("[func:clearDb] ", "clearDb Finish ")
}

func StartAlert() {
	ClearDb()
	seelog.Info("[func:StartAlert] ", "starting run AlertCheck ")
	timeStartStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
	dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
	sql := `CREATE TABLE IF NOT EXISTS [alertlog-` + dateStartStr + `] (
			logtime   VARCHAR (8),
				fromname  VARCHAR (15),
				toname    VARCHAR (15),
				alerttype INT (1)
	);`
	for _, v := range g.Cfg.Targets {
		if v.Addr != g.Cfg.Ip {
			checktimeStartStr := time.Unix((time.Now().Unix() - int64(v.Thdchecksec)), 0).Format("2006-01-02 15:04")
			g.DLock.Lock()
			sql := "SELECT ifnull(max(avgdelay),0) maxavgdelay, ifnull(max(losspk),0) maxlosspk ,count(1) Cnt FROM  `pinglog-" + v.Addr + "` where lastcheck > '" + checktimeStartStr + "' and (cast(avgdelay as double) > " + strconv.Itoa(v.Thdavgdelay) + " or cast(losspk as double) > " + strconv.Itoa(v.Thdloss) + ") "
			rows, err := g.Db.Query(sql)
			seelog.Debug("[func:StartAlert] ", sql)
			if err != nil {
				seelog.Error("[func:StartAlert] ", err)
				return
			}
			for rows.Next() {
				l := new(g.TopoLog)
				err := rows.Scan(&l.Maxavgdelay, &l.Maxlosspk, &l.Cnt)
				if err != nil {
					seelog.Error("[func:StartAlert] ", err)
				}
				sec, _ := strconv.Atoi(l.Cnt)
				if sec >= v.Thdoccnum {
					sql = sql + "insert into [alertlog-" + dateStartStr + "] (logtime,fromname,toname,alerttype) values('" + timeStartStr + "','" + g.Cfg.Name + "','" + v.Name + "','1');"
				}
			}
			rows.Close()
			g.DLock.Unlock()
		}
	}
	SqlExec(sql)
	seelog.Info("[func:StartAlert] ", "AlertCheck finish ")
}
