package funcs

import (
	"../ping"
	"../g"
	"database/sql"
	"github.com/gy-games-libs/seelog"
)


func StartSysPing(t g.Target, db *sql.DB,config g.Config){
	seelog.Info("Start SysPing "+t.Addr+"..")
	rt := ping.SysPing(t.Addr)
	StoragePing(rt,t,db)
	seelog.Info("Finish SysPing "+t.Addr+"..")
}

func StartGoPing(t g.Target, db *sql.DB,config g.Config){
	seelog.Info("Start GoPing "+t.Addr+"..")
	rt := ping.GoPing(t.Addr)
	StoragePing(rt,t,db)
	seelog.Info("Finish GoPing "+t.Addr+"..")
}

func StartFPing(db *sql.DB,config g.Config){

}