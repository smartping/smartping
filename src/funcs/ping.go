package funcs

import (
	"../g"
	"../ping"
	"database/sql"
	"github.com/gy-games-libs/seelog"
)

func StartPing(t g.Target, db *sql.DB, config g.Config) {
	seelog.Info("Start SysPing " + t.Addr + "..")
	rt := ping.IcmpPing(t.Addr)
	StoragePing(rt, t, db)
	seelog.Info("Finish SysPing " + t.Addr + "..")
}
