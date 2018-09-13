package funcs

import (
	"github.com/gy-games/smartping/src/g"
	_ "github.com/mattn/go-sqlite3"
	"github.com/cihub/seelog"
	"time"
)

//sql exec function
func SqlExec(sql string) {
	seelog.Debug("[func:StartPing] ", sql)
	g.DLock.Lock()
	g.Db.Exec(sql)
	g.DLock.Unlock()
}

//create ping database table
func CreatePingTable(t g.Target) {
	seelog.Info("[func:CreatePingTable] CreateDB `pinglog-", t.Addr, "` Start..")
	sql := `CREATE TABLE IF NOT EXISTS [pinglog-` + t.Addr + `] (
	    logtime   VARCHAR (8),
	    maxdelay  VARCHAR (3),
	    mindelay  VARCHAR (3),
	    avgdelay  VARCHAR (3),
	    sendpk    VARCHAR (2),
	    revcpk    VARCHAR (2),
	    losspk    VARCHAR (3),
	    lastcheck VARCHAR (16),
	    PRIMARY KEY (
		logtime
	    )
	);
	CREATE INDEX  IF NOT EXISTS  "lc" ON [pinglog-` + t.Addr + `] (
	    lastcheck
	);`
	SqlExec(sql)
	seelog.Info("[func:CreatePingTable] CreateDB `pinglog-", t.Addr, "` Finish..")
}

//clear timeout alert table
func ClearAlertTable() {
	seelog.Info("[func:ClearAlertTable] ", "starting run ClearAlertTable ")
	sql := ""
	reminList := map[string]bool{}
	for i := 0; i < g.Cfg.Alerthistory; i++ {
		reminList["alertlog-"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")] = true
	}
	listpreout := []string{}
	g.DLock.Lock()
	querySQl := "SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'"
	rows, err := g.Db.Query(querySQl)
	g.DLock.Unlock()
	if err != nil {
		seelog.Error("[funcs:ClearAlertTable] Query ", err)
		return
	}
	for rows.Next() {
		var l string
		err := rows.Scan(&l)
		if err != nil {
			seelog.Error("[funcs:ClearAlertTable] Rows ", err)
			return
		}
		listpreout = append(listpreout, l)
	}
	rows.Close()
	for _, v := range listpreout {
		if _, ok := reminList[v]; !ok {
			sql = sql + "DROP TABLE [" + v + "];"
		}
	}
	SqlExec(sql)
	seelog.Info("[func:ClearAlertTable] ", "ClearAlertTable Finish ")
}

//clear unused ping table
func ClearPingTable() {
	seelog.Info("[func:ClearPingTable] ", "ClearPingTable Finish ")
	reminList := map[string]bool{}
	for _, target := range g.Cfg.Targets {
		reminList["pinglog-"+target.Addr] = true
	}
	sql := ""
	listpreout := []string{}
	g.DLock.Lock()
	lrows, lerr := g.Db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%pinglog%'")
	g.DLock.Unlock()
	if lerr != nil {
		seelog.Error("[funcs:ClearPingTable] Query ", lerr)
		return
	}
	for lrows.Next() {
		var l string
		err := lrows.Scan(&l)
		if err != nil {
			seelog.Error("funcs:ClearPingTable] Rows ", err)
			return
		}
		listpreout = append(listpreout, l)
	}
	lrows.Close()
	for _, v := range listpreout {
		if _, ok := reminList[v]; !ok {
			sql = sql + "DROP TABLE [" + v + "];"
		}
	}
	SqlExec(sql)
	seelog.Info("[func:ClearPingTable] ", "ClearPingTable Finish ")
}
