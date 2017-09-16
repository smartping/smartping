package funcs

import (
	"../g"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/gy-games-libs/go-sqlite3"
	"github.com/gy-games-libs/resty"
	"log"
	"strconv"
	"time"
)

func StartPing(t g.Target, db *sql.DB) {
	log.Println("starting runPingTest ", t.Name)
	pingres := Ping(t.Addr, "10")
	//log.Print(pingres)
	//status := g.PingResult{SendPk: pingres.SendPk, RevcPk: pingres.RevcPk, LossPk: pingres.LossPk, MaxDelay: pingres.MaxDelay, MinDelay: pingres.MinDelay, AvgDelay: pingres.AvgDelay, LastCheck: time.Now().Format("2006-01-02 15:04")}
	g.DLock.Lock()
	//create new table
	db.Exec(`CREATE TABLE IF NOT EXISTS [pinglog-` + t.Addr + `] (
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
	CREATE INDEX "lc" ON [pinglog-` + t.Addr + `] (
	    lastcheck
	);`)
	stmt, _ := db.Prepare("REPLACE INTO [pinglog-" + t.Addr + "] (logtime, maxdelay, mindelay, avgdelay, sendpk, revcpk, losspk, lastcheck) values(?,?,?,?,?,?,?,?)")
	stmt.Exec(time.Now().Format("02 15:04"), pingres.MaxDelay, pingres.MinDelay, pingres.AvgDelay, pingres.SendPk, pingres.RevcPk, pingres.LossPk, time.Now().Format("2006-01-02 15:04"))
	//log.Print(pingres)
	stmt.Close()
	g.DLock.Unlock()
	log.Printf("PingTest on %s finish!", t.Name)
}

func StartAlert(config g.Config, db *sql.DB) {
	log.Println("starting run AlertCheck ")
	timeStartStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
	dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
	sql := `CREATE TABLE IF NOT EXISTS [alertlog-` + dateStartStr + `] (
			logtime   VARCHAR (8),
				fromname  VARCHAR (15),
				toname    VARCHAR (15),
				alerttype INT (1)
		);`
	reminList := map[string]bool{}
	for i := 0; i < config.Alerthistory; i++ {
		reminList["alertlog-"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")] = true
	}
	listpreout := []string{}
	lrows, lerr := db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'")
	if lerr == nil {
		for lrows.Next() {
			var l string
			err := lrows.Scan(&l)
			if err != nil {
				fmt.Println(err)
			}
			listpreout = append(listpreout, l)
		}
		lrows.Close()
	}
	for _, v := range listpreout {
		if _, ok := reminList[v]; !ok {
			sql = sql + "DROP TABLE [" + v + "];"
		}
	}
	sec, _ := strconv.Atoi(config.Timeout)
	resp, _ := resty.SetTimeout(time.Second * time.Duration(sec)).R().Get("http://127.0.0.1:" + strconv.Itoa(config.Port) + "/api/topology.json")
	if resp.StatusCode() == 200 {
		pingstatus := make(map[string]interface{})
		json.Unmarshal([]byte(resp.String()), &pingstatus)
		for target, value := range pingstatus {
			if value != "true" {
				sql = sql + "insert into [alertlog-" + dateStartStr + "] (logtime,fromname,toname,alerttype) values('" + timeStartStr + "','" + config.Name + "','" + target + "','1');"
			}
		}
	} else {
		sql = sql + "insert into [alertlog-" + dateStartStr + "] (logtime,fromname,toname,alerttype) values('" + timeStartStr + "','" + config.Name + "','" + config.Name + "','" + config.Name + "','2');"
	}
	g.DLock.Lock()
	db.Exec(sql)
	g.DLock.Unlock()
	log.Println("AlertCheck finish")
}

/*
func StartAlertGlobal(config g.Config, db *sql.DB){

	//for {
		log.Println("starting run AlertCheck ")
		timeStartStr := time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04")
		dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
		sql := `CREATE TABLE IF NOT EXISTS [alertlog-`+dateStartStr+`] (
			logtime   VARCHAR (8),
				fromname  VARCHAR (15),
				toname    VARCHAR (15),
				alerttype INT (1)
		);`
		reminList := map[string]bool{}
		for i:=0;i<config.Alerthistory;i++ {
			reminList["alertlog-"+time.Unix((time.Now().Unix()-int64(86400*i)), 0).Format("20060102")]=true
		}
		listpreout := []string{}
		lrows, lerr := db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'")
		if lerr==nil{
			for lrows.Next() {
				var l string
				err := lrows.Scan(&l,)
				if err != nil {
					fmt.Println(err)
				}
				listpreout = append(listpreout,l)
			}
			lrows.Close()
		}
		for _,v:= range listpreout{
			if _, ok := reminList[v]; !ok {
				sql = sql +"DROP TABLE ["+v+"];";
			}
		}
		for _, v := range config.Targets {
			sec, _ := strconv.Atoi(config.Timeout)
			resp, _ := resty.SetTimeout(time.Second * time.Duration(sec)).R().Get("http://" + v.Addr + ":" + strconv.Itoa(config.Port) + "/api/topology.json")
			if resp.StatusCode() == 200 {
				pingstatus := make(map[string]interface{})
				json.Unmarshal([]byte(resp.String()), &pingstatus)
				for target, value := range pingstatus {
					if value != "true" {
						sql = sql + "insert into [alertlog-"+dateStartStr+"] (logtime,fromname,toname,alerttype) values('" + timeStartStr + "','" + v.Name + "','" + target + "','1');"
					}
				}
			} else {
				sql = sql + "insert into [alertlog-"+dateStartStr+"] (logtime,fromname,toname,alerttype) values('" + timeStartStr + "','" + config.Name + "','" + v.Name + "','" + v.Name + "','2');"
			}
		}
		g.DLock.Lock()
		db.Exec(sql)
		g.DLock.Unlock()
		log.Print(sql)
		log.Println("AlertCheck finish")
		time.Sleep(time.Duration(config.Alertcycle) * 60 * time.Second)
	//}
}
*/
