package funcs

import (
	"github.com/cihub/seelog"
	"time"
	"github.com/smartping/smartping/src/g"
	"io/ioutil"
	"strings"
	"os"
)

//clear timeout alert table
func ClearArchive() {
	seelog.Info("[func:ClearArchive] ", "starting run ClearArchive ")
	reminPingList := map[string]bool{}
	for _, target := range g.Cfg.Targets {
		reminPingList["ping_"+target.Addr] = true
	}
	seelog.Debug("[func:ClearArchive] PingLog DB Remin List",reminPingList)
	removeFile("ping",reminPingList)
	reminAlertList := map[string]bool{}
	for i := 0; i < g.Cfg.Alerthistory; i++ {
		reminAlertList["alert_"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")] = true
	}
	seelog.Debug("[func:ClearArchive] Alert DB Remin List",reminAlertList)
	removeFile("alert",reminAlertList)
	seelog.Info("[func:ClearArchive] ", "ClearArchive Finish ")
}

func removeFile(t string,reminList map[string]bool){
	allPingList, err := ioutil.ReadDir(g.Root+"/db/"+t+"/")
	if err!=nil{
		seelog.Error("[func:ClearBucket] Get "+t+" db list error", err)
	}
	for _, dbfile := range allPingList {
		if strings.Contains(dbfile.Name(),".db") && ! strings.Contains(dbfile.Name(),".db.lock"){
			dbname := t+"_"+strings.Split(dbfile.Name(),".db")[0]
			if ok,_:=reminList[dbname];!ok{
				//if _,ok := g.DbMap[dbname]; ok {
				//	g.DbMap[dbname].Close()
				//}
				db,ok := g.DbMap.Get(dbname)
				if ok==nil{
					db.Close()
				}
				seelog.Debug("[func:removeFile] ",g.Root+"/db/"+t+"/"+dbfile.Name())
				os.Remove(g.Root+"/db/"+t+"/"+dbfile.Name())
				os.Remove(g.Root+"/db/"+t+"/"+dbfile.Name()+".lock")
			}
		}
	}
}