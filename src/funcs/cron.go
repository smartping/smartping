package funcs

import (
	"log"
	"time"
	//"../cmdping"
	"../g"
	"database/sql"
	_"github.com/gy-games-libs/go-sqlite3"
)

func StartPing(db *sql.DB , c g.Config, t g.Target, res chan g.TargetStatus) {
	go runPingTest(db, c, t, res)
}

func runPingTest( db *sql.DB, c g.Config, t g.Target, res chan g.TargetStatus) {
	for {
		log.Println("starting runPingTest ", t.Name)
		var status g.TargetStatus
		pingres := Ping(t.Addr, t.Addr, t.Interval)
		lastcheck :=time.Now().Format("2006-01-02 15:04")
		logtime :=time.Now().Format("02 15:04")
		status = g.TargetStatus{Target: &t, SendPk: pingres.SendPk, RevcPk: pingres.RevcPk, LossPk: pingres.LossPk, MaxDelay: pingres.MaxDelay, MinDelay: pingres.MinDelay, AvgDelay: pingres.AvgDelay, LastCheck: lastcheck }
		//Insert data
		g.DLock.Lock()
		log.Println("REPLACE INTO pinglog ",logtime, t.Addr,t.Name, pingres.MaxDelay, pingres.AvgDelay, pingres.MinDelay, pingres.SendPk, pingres.RevcPk, pingres.LossPk, lastcheck)
		stmt, _ := db.Prepare("REPLACE INTO pinglog(logtime, ip, name, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk, lastcheck) values(?,?,?,?,?,?,?,?,?,?)")
		stmt.Exec(logtime, t.Addr,t.Name, pingres.MaxDelay, pingres.AvgDelay, pingres.MinDelay, pingres.SendPk, pingres.RevcPk, pingres.LossPk, lastcheck )
		stmt.Close()
		g.DLock.Unlock()
		res <- status
		log.Printf("runPingTest on %s finish!", t.Name)
	}

}
