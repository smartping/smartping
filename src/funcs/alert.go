package funcs

import (
	"bytes"
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"time"
	"github.com/boltdb/bolt"
	"encoding/json"
	"strings"
	"strconv"
	"fmt"
)

//alert main function
func StartAlert() {
	seelog.Info("[func:StartAlert] ", "starting run AlertCheck ")
	dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
	for _, v := range g.Cfg.Targets {
		if v.Addr != g.Cfg.Ip {
			s := CheckAlertStatus(v)
			if s=="true"{
				g.AlertStatus[v.Addr]=true
			}
			_, haskey := g.AlertStatus[v.Addr];
			if ( !haskey && s=="false" ) || ( s=="false" && g.AlertStatus[v.Addr] ) {
				seelog.Debug("[func:StartAlert] ",v.Addr+" Alert!")
				g.AlertStatus[v.Addr]=false
				l := g.AlertLog{}
				l.Logtime = time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
				l.Toname = v.Name
				l.Fromname = g.Cfg.Name
				tracrtString := ""
				hops, err := nettools.RunTrace(v.Addr, time.Second, 64, 6)
				if nil != err {
					seelog.Error("[func:StartAlert] Traceroute error ", err)
					tracrtString = err.Error()
				} else {
					tracrt := bytes.NewBufferString("")
					for i, hop := range hops {
						if hop.Addr == nil {
							fmt.Fprintf(tracrt, "* * *\n")
						} else {
							fmt.Fprintf(tracrt, "%d (%s) %v %v %v\n", i+1, hop.Addr, hop.MaxRTT, hop.AvgRTT, hop.MinRTT)
						}
					}
					tracrtString = tracrt.String()
				}
				l.Tracert = tracrtString
				db:=g.GetDb("alert",dateStartStr)
				err = db.Update(func(tx *bolt.Tx) error {
					b, err := tx.CreateBucketIfNotExists([]byte("alertlog"))
					if err != nil {
						return fmt.Errorf("create bucket error : %s", err)
					}
					jdata,_ :=json.Marshal(l)
					err = b.Put([]byte(l.Logtime+v.Name), []byte(string(jdata)))
					if err != nil {
						return fmt.Errorf("put data error: %s", err)
					}
					return nil
				})
				if err != nil {
					seelog.Error("[func:StartAlert] Data Storage Error: ",err)
				}
			}

		}
	}
	seelog.Info("[func:StartAlert] ", "AlertCheck finish ")
}

func CheckAlertStatus(v g.Target) string{
	timeStartStr := time.Unix((time.Now().Unix() - int64(v.Thdchecksec)), 0).Format("2006-01-02 15:04")
	timeEndStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
	result := "false"
	db:=g.GetDb("ping",v.Addr)
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pinglog"))
		if b==nil{
			result = "unknown"
			return nil
		}
		c:= b.Cursor()
		min := []byte(timeStartStr[8:])
		max := []byte(timeEndStr[8:])
		ectime:=0
		for k, val := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, val = c.Next() {
			l := new(g.PingLog)
			err := json.Unmarshal(val,&l)
			if err!=nil{
				continue
			}
			if l.Logtime > timeStartStr && l.Logtime < timeEndStr {
				seelog.Debug(v.Name + "|" + strings.Split(l.Avgdelay, ".")[0] + "|" + strconv.Itoa(v.Thdavgdelay) + "|" + strings.Split(l.Losspk, ".")[0] + "|" + strconv.Itoa(v.Thdloss))
				avgdelay, _ := strconv.Atoi(strings.Split(l.Avgdelay, ".")[0])
				losspk, _ := strconv.Atoi(strings.Split(l.Losspk, ".")[0])
				if avgdelay > v.Thdavgdelay || losspk >= v.Thdloss {
					ectime = ectime + 1
				}
			}
		}
		if ectime<v.Thdoccnum{
			result =  "true"
		}
		return nil
	})
	return result
}
