package funcs

import (
	"github.com/cihub/seelog"
	"smartping/src/g"
	"strconv"
)

//clear timeout alert table
func ClearArchive() {
	seelog.Info("[func:ClearArchive] ", "starting run ClearArchive ")
	g.DLock.Lock()
	g.Db.Exec("delete from alertlog where logtime < date('now','start of day','-" + strconv.Itoa(g.Cfg.Base["Archive"]) + " day')")
	g.Db.Exec("delete from mappinglog where logtime < date('now','start of day','-" + strconv.Itoa(g.Cfg.Base["Archive"]) + " day')")
	g.Db.Exec("delete from pinglog where logtime < date('now','start of day','-" + strconv.Itoa(g.Cfg.Base["Archive"]) + " day')")
	g.DLock.Unlock()
	seelog.Info("[func:ClearArchive] ", "ClearArchive Finish ")
}
