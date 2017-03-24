package main

import (
	"log"
	"time"
	"smartping/cmdping"
	"database/sql"
	_"github.com/mattn/go-sqlite3"
)

func startPing(db *sql.DB , c Config, t Target, res chan TargetStatus) {
	go runPingTest(db, c, t, res)
}

func runPingTest( db *sql.DB, c Config, t Target, res chan TargetStatus) {
	for {
		log.Println("starting runPingTest ", t.Name)
		var status TargetStatus
		pingres := cmdping.Ping(t.Addr, t.Addr, t.Interval)
		lastcheck :=time.Now().Format("2006-01-02 15:04")
		logtime :=time.Now().Format("02 15:04")
		status = TargetStatus{Target: &t, SendPk: pingres.SendPk, RevcPk: pingres.RevcPk, LossPk: pingres.LossPk, MaxDelay: pingres.MaxDelay, MinDelay: pingres.MinDelay, AvgDelay: pingres.AvgDelay, LastCheck: lastcheck }
		//Insert data
		lock.Lock()
		log.Println("INSERT ",lastcheck, t.Addr,t.Name, pingres.MaxDelay, pingres.AvgDelay, pingres.MinDelay, pingres.SendPk, pingres.RevcPk, pingres.LossPk)
		stmt, _ := db.Prepare("REPLACE INTO pinglog(logtime, ip, name, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk, lastcheck) values(?,?,?,?,?,?,?,?,?,?)")
		stmt.Exec(logtime, t.Addr,t.Name, pingres.MaxDelay, pingres.AvgDelay, pingres.MinDelay, pingres.SendPk, pingres.RevcPk, pingres.LossPk, lastcheck )
		lock.Unlock()
		res <- status
		log.Printf("runPingTest on %s finish!", t.Name)
	}

}
