package funcs

import (
	"database/sql"
	"github.com/gy-games-libs/seelog"
	"time"
	"../g"
	_ "github.com/gy-games-libs/go-sqlite3"
)

func CreateDB(t g.Target, db *sql.DB) {
	seelog.Info("[func:CreateDB] CreateDB `pinglog-",t.Addr,"` Start..")
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
	);
	`
	seelog.Debug("[func:CreateDB] ", sql)
	g.DLock.Lock()
	db.Exec(sql)
	g.DLock.Unlock()
	seelog.Info("[func:CreateDB] CreateDB `pinglog-",t.Addr,"` Finish..")
}


func StoragePing(pingres g.PingSt,t g.Target, db *sql.DB) {
	logtime := time.Now().Format("02 15:04")
	checktime := time.Now().Format("2006-01-02 15:04")
	seelog.Info("[func:StartPing] ", "(", checktime, ")Starting runPingTest ", t.Name)
	//pingres := Ping(t.Addr)
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
	);
	`
	sql = sql + "REPLACE INTO [pinglog-" + t.Addr + "] (logtime, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk, lastcheck) values('" + logtime + "','" + pingres.MaxDelay + "','" + pingres.MinDelay + "','" + pingres.AvgDelay + "','" + pingres.SendPk + "','" + pingres.RevcPk + "','" + pingres.LossPk + "','" + checktime + "')"
	seelog.Debug("[func:StartPing] ", sql)
	g.DLock.Lock()
	db.Exec(sql)
	g.DLock.Unlock()
	seelog.Info("[func:StartPing] ", "(", checktime, ") PingTest on ", t.Name, " finish!")
}

