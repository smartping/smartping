package funcs

import (
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"time"
	"github.com/boltdb/bolt"
	"bytes"
	"strconv"
)

//clear timeout alert table
func ClearBucket() {
	seelog.Info("[func:ClearBucket] ", "starting run ClearBucket ")
	reminList := map[string]bool{}
	for i := 0; i < g.Cfg.Alerthistory; i++ {
		reminList["alertlog-"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")] = true
	}
	for _, target := range g.Cfg.Targets {
		reminList["pinglog-"+target.Addr] = true
	}
	allBucket :=map[string]bool{}
	err := g.Db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			allBucket[string(name)]=true
			return nil
		})
	})
	if err != nil {
		seelog.Error("[func:ClearBucket ] Find Error ",err)
	}
	seelog.Debug("[func:ClearBucket ] reminList ",reminList)
	seelog.Debug("[func:ClearBucket ] allBucket ",allBucket)
	for k,_ :=range allBucket{
		if ok,_:=reminList[k];!ok{
			err = g.Db.Update(func(tx *bolt.Tx) error {
				tx.DeleteBucket([]byte(k))
				seelog.Debug("[func:ClearBucket ] DeleteBucket ",k)
				return nil
			})
			if err!=nil{
				seelog.Error("[func:ClearBucket ] Delete Error ",err)
			}
		}
	}
	seelog.Info("[func:ClearBucket] ", "ClearBucket Finish ")
}

//clear unused ping table
func ClearPingLog() {
	seelog.Info("[func:ClearPingLog] ", "ClearPingLog Finish ")
	for _, target := range g.Cfg.Targets {
		if g.Cfg.Ip != target.Addr{
			preDelkeyList := []string{}
			err := g.Db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(target.Addr))
				if b==nil{
					return nil
				}
				archiveTime := time.Now().Unix() - int64(g.Cfg.Alerthistory * 86400)
				archiveTimeStr := time.Unix(archiveTime, 0).Format("2006-01-02 15:04")
				min := []byte("0000-00-00 00:00")
				max := []byte(archiveTimeStr)
				c := b.Cursor()
				for k, _ := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, _ = c.Next()  {
					preDelkeyList = append(preDelkeyList,string(k))
				}
				return nil
			})
			if err!=nil{
				seelog.Error("[func:ClearPingLog ] Find Keys Error ",err)
			}
			if len(preDelkeyList)>0{
				err = g.Db.Update(func(tx *bolt.Tx) error {
					b := tx.Bucket([]byte(target.Addr))
					for k,_ := range preDelkeyList{
						b.Delete([]byte(strconv.Itoa(k)))
					}
					return nil
				})
				if err!=nil {
					seelog.Error("[func:ClearPingLog ] Delete Keys Error ", err)
				}
				seelog.Debug("[func:ClearPingLog ] DELETE KEY : ",len(preDelkeyList))
			}
		}
	}
	seelog.Info("[func:ClearPingLog] ", "ClearPingLog Finish ")
}