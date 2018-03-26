package funcs

import (
	"../g"
	_ "github.com/gy-games-libs/go-sqlite3"
	"github.com/gy-games-libs/seelog"
	"strconv"
	"time"
)

func SqlExec(sql string) {
	seelog.Debug("[func:StartPing] ", sql)
	g.DLock.Lock()
	g.Db.Exec(sql)
	g.DLock.Unlock()
}

func CreateDB(t g.Target) {
	seelog.Info("[func:CreateDB] CreateDB `pinglog-", t.Addr, "` Start..")
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
	seelog.Info("[func:CreateDB] CreateDB `pinglog-", t.Addr, "` Finish..")
}

func StoragePing(pingres g.PingSt, t g.Target) {
	logtime := time.Now().Format("02 15:04")
	checktime := time.Now().Format("2006-01-02 15:04")
	seelog.Info("[func:StartPing] ", "(", checktime, ")Starting runPingTest ", t.Name)
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
	sql = sql + "REPLACE INTO [pinglog-" + t.Addr + "] (logtime, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk, lastcheck) values('" + logtime + "','" + strconv.FormatFloat(pingres.MaxDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.MinDelay, 'f', 2, 64) + "','" + strconv.FormatFloat(pingres.AvgDelay, 'f', 2, 64) + "','" + strconv.Itoa(pingres.SendPk) + "','" + strconv.Itoa(pingres.RevcPk) + "','" + strconv.Itoa(pingres.LossPk) + "','" + checktime + "')"
	SqlExec(sql)
	seelog.Info("[func:StartPing] ", "(", checktime, ") PingTest on ", t.Name, " finish!")
}
