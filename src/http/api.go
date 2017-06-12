package http

import (
	"net/http"
	"time"
	"fmt"
	"encoding/json"
	"bytes"
	"strconv"
	"log"
	"database/sql"
	"../g"
)

func configApiRoutes(port int, state *g.State ,db *sql.DB ,config g.Config){

	//graph data api
	http.HandleFunc("/api/status.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		var where string
		var timeStart int64
		var timeEnd int64
		var timeStartStr string
		var timeEndStr string
		if len(r.Form["starttime"]) > 0 && len(r.Form["endtime"]) > 0 {
			timeStartStr = r.Form["starttime"][0]
			timeEndStr = r.Form["endtime"][0]
			tms, _ := time.Parse("2006-01-02 15:04", timeStartStr)
			timeStart = tms.Unix()-8*60*60
			tmn, _ := time.Parse("2006-01-02 15:04", timeEndStr)
			timeEnd = tmn.Unix()-8*60*60
		}else{
			timeStart = time.Now().Unix()-2*60*60
			timeEnd = time.Now().Unix()
			timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			timeEndStr = time.Unix(timeEnd, 0).Format("2006-01-02 15:04")
		}
		cnt := int((timeEnd-timeStart)/60)
		var lastcheck []string
		var ip []string
		var name []string
		var maxdelay []string
		var mindelay []string
		var avgdelay []string
		var sendpk []string
		var revcpk []string
		var losspk []string
		for i:=0;i<cnt;i++ {
			lastcheck = append(lastcheck,time.Unix(timeStart, 0).Format("2006-01-02 15:04"))
			ip        = append(ip,"0")
			name      = append(name,"0")
			maxdelay  = append(maxdelay,"0")
			mindelay  = append(mindelay,"0")
			avgdelay  = append(avgdelay,"0")
			sendpk    = append(sendpk,"0")
			revcpk    = append(revcpk,"0")
			losspk    = append(losspk,"0")
			timeStart = timeStart+60
		}
		if len(r.Form["ip"]) > 0 {
			where = where + "and ip = '"+ r.Form["ip"][0] + "'"
		}
		if len(r.Form["name"]) > 0 {
			where = where + "and name = '"+ r.Form["name"][0] + "'"
		}
		g.DLock.Lock()
		rows, _ := db.Query("SELECT * FROM pinglog where 1=1 and lastcheck between '"+timeStartStr+"' and '"+timeEndStr+"' "+where+"")
		//foreach all data
		for rows.Next() {
			l := new(g.LogInfo)
			err := rows.Scan(&l.Logtime, &l.Ip, &l.Name, &l.Maxdelay, &l.Mindelay, &l.Avgdelay, &l.Sendpk, &l.Revcpk, &l.Losspk, &l.Lastcheck,)
			if err != nil {
				fmt.Println(err)
			}
			for n, v := range lastcheck{
				if v==l.Lastcheck{
					maxdelay[n] = l.Maxdelay
					mindelay[n] = l.Mindelay
					avgdelay[n] = l.Avgdelay
					losspk[n] = l.Losspk
				}
			}
		}
		g.DLock.Unlock()
		preout := map[string][]string{
			"lastcheck": lastcheck,
			"ip": ip,
			"name": name,
			"maxdelay": maxdelay,
			"mindelay": mindelay,
			"avgdelay": avgdelay,
			"sendpk": sendpk,
			"revcpk": revcpk,
			"losspk": losspk,
		}
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
	})

	//config api
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		config, _ := json.Marshal(config)
		var out bytes.Buffer
		json.Indent(&out, config, "", "\t")
		o := out.String()
		fmt.Fprintln(w, o)
	})

	//Topology alert data api
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		state.Lock.Lock()
		defer state.Lock.Unlock()
		preout := make(map[string]string)
		var 		timeStart int64
		var 		timeStartStr string
		sec 		:= config.Thdchecksec
		timeStart   	= time.Now().Unix()-int64(sec)
		timeStartStr 	= time.Unix(timeStart, 0).Format("2006-01-02 15:04")
		Thdloss     	:= config.Thdloss
		Thdavgdelay 	:= config.Thdavgdelay
		Thdoccnum   	:= config.Thdoccnum
		dbrst   	:= map[string]int{}
		g.DLock.Lock()
		rows, _ := db.Query("SELECT ip,name,max(avgdelay) maxavgdelay, max(losspk) maxlosspk ,count(1) Cnt FROM  pinglog where lastcheck > '"+timeStartStr+"' and (cast(avgdelay as double) > "+strconv.Itoa(Thdavgdelay)+" or cast(losspk as double) > "+strconv.Itoa(Thdloss)+") group by ip,name ")
		//log.Print("SELECT ip,name,max(avgdelay) maxavgdelay, max(losspk) maxlosspk ,count(1) Cnt FROM  pinglog where lastcheck > '"+timeStartStr+"' and (cast(avgdelay as double) > "+strconv.Itoa(Thdavgdelay)+" or cast(losspk as double) > "+strconv.Itoa(Thdloss)+") group by ip,name ")
		for rows.Next() {
			l 		:= new( g.TopoLog )
			rows.Scan( &l.Ip, &l.Name,&l.Maxavgdelay,&l.Maxlosspk,&l.Cnt,)
			log.Print(l)
			dbrst[l.Ip],_ = strconv.Atoi(l.Cnt)
		}
		g.DLock.Unlock()
		for _,v := range state.State{
			for _,t:=range config.Targets {
				if t.Name==v.Target.Name{
					preout[v.Target.Name]="true"
					if (t.Thdloss!=0 && t.Thdloss != Thdloss) || (t.Thdoccnum != 0 && t.Thdoccnum != Thdoccnum) || (t.Thdavgdelay != 0 && t.Thdavgdelay != Thdavgdelay ){
						g.DLock.Lock()
						rows, _ := db.Query("SELECT ip,name,max(avgdelay) maxavgdelay, max(losspk) maxlosspk ,count(1) Cnt FROM  pinglog where lastcheck > '"+timeStartStr+"' and (cast(avgdelay as double) > "+strconv.Itoa(t.Thdavgdelay)+" or cast(losspk as double) > "+strconv.Itoa(t.Thdloss)+") and ip = '"+v.Target.Addr+"' ")
						//log.Print("SELECT ip,name,max(avgdelay) maxavgdelay, max(losspk) maxlosspk ,count(1) Cnt FROM  pinglog where lastcheck > '"+timeStartStr+"' and (cast(avgdelay as double) > "+strconv.Itoa(t.Thdavgdelay)+" or cast(losspk as double) > "+strconv.Itoa(t.Thdloss)+") and ip = '"+v.Target.Addr+"' ")
						for rows.Next() {
							l := new( g.TopoLog )
							rows.Scan( &l.Ip, &l.Name,&l.Maxavgdelay,&l.Maxlosspk,&l.Cnt,)
							dbrst[l.Ip],_ = strconv.Atoi(l.Cnt)
						}
						g.DLock.Unlock()
						if( dbrst[v.Target.Addr]> t.Thdoccnum ){
							preout[v.Target.Name]="false"
						}
					}else{
						//log.Print(v.Target.Name+": DB TH CNT:"+strconv.Itoa(dbrst[v.Target.Addr])+" Thdiccnun:"+strconv.Itoa(Thdoccnum))
						if( dbrst[v.Target.Addr]> Thdoccnum ){
							preout[v.Target.Name]="false"
						}
					}
				}
			}
		}
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
	})

}
