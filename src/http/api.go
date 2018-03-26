package http

import (
	"../funcs"
	"../g"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func configApiRoutes() {

	//config api
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		nconf := g.Config{}
		nconf = g.Cfg
		nconf.Password = ""
		onconf, _ := json.Marshal(nconf)
		var out bytes.Buffer
		json.Indent(&out, onconf, "", "\t")
		o := out.String()
		fmt.Fprintln(w, o)
	})

	//graph data api
	http.HandleFunc("/api/ping.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		var tableip string
		var where string
		var timeStart int64
		var timeEnd int64
		var timeStartStr string
		var timeEndStr string
		if len(r.Form["starttime"]) > 0 && len(r.Form["endtime"]) > 0 {
			timeStartStr = r.Form["starttime"][0]
			if timeStartStr != "" {
				tms, _ := time.Parse("2006-01-02 15:04", timeStartStr)
				timeStart = tms.Unix() - 8*60*60
			} else {
				timeStart = time.Now().Unix() - 2*60*60
				timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			}
			timeEndStr = r.Form["endtime"][0]
			if timeEndStr != "" {
				tmn, _ := time.Parse("2006-01-02 15:04", timeEndStr)
				timeEnd = tmn.Unix() - 8*60*60
			} else {
				timeEnd = time.Now().Unix()
				timeEndStr = time.Unix(timeEnd, 0).Format("2006-01-02 15:04")
			}
		} else {
			timeStart = time.Now().Unix() - 2*60*60
			timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			timeEnd = time.Now().Unix()
			timeEndStr = time.Unix(timeEnd, 0).Format("2006-01-02 15:04")
		}
		cnt := int((timeEnd - timeStart) / 60)
		var lastcheck []string
		var maxdelay []string
		var mindelay []string
		var avgdelay []string
		var sendpk []string
		var revcpk []string
		var losspk []string
		for i := 0; i < cnt+1; i++ {
			lastcheck = append(lastcheck, time.Unix(timeStart, 0).Format("2006-01-02 15:04"))
			maxdelay = append(maxdelay, "0")
			mindelay = append(mindelay, "0")
			avgdelay = append(avgdelay, "0")
			sendpk = append(sendpk, "0")
			revcpk = append(revcpk, "0")
			losspk = append(losspk, "0")
			timeStart = timeStart + 60
		}
		if len(r.Form["ip"]) > 0 {
			tableip = r.Form["ip"][0]
		} else {
			tableip = ""
		}
		g.DLock.Lock()
		sql := "SELECT logtime,maxdelay,mindelay,avgdelay,sendpk,revcpk,losspk,lastcheck FROM `pinglog-" + tableip + "` where 1=1 and lastcheck between '" + timeStartStr + "' and '" + timeEndStr + "' " + where + ""
		rows, err := g.Db.Query(sql)
		seelog.Debug("[func:/api/ping.json] ", sql)
		if err == nil {
			for rows.Next() {
				l := new(g.LogInfo)
				err := rows.Scan(&l.Logtime, &l.Maxdelay, &l.Mindelay, &l.Avgdelay, &l.Sendpk, &l.Revcpk, &l.Losspk, &l.Lastcheck)
				if err != nil {
					seelog.Error("[/api/ping.json] ", err)
				}
				for n, v := range lastcheck {
					if v == l.Lastcheck {
						maxdelay[n] = l.Maxdelay
						mindelay[n] = l.Mindelay
						avgdelay[n] = l.Avgdelay
						losspk[n] = l.Losspk
						sendpk[n] = l.Sendpk
						revcpk[n] = l.Revcpk
					}
				}
			}
			rows.Close()
		}
		g.DLock.Unlock()
		preout := map[string][]string{
			"lastcheck": lastcheck,
			"maxdelay":  maxdelay,
			"mindelay":  mindelay,
			"avgdelay":  avgdelay,
			"sendpk":    sendpk,
			"revcpk":    revcpk,
			"losspk":    losspk,
		}
		RenderJson(w, preout)
		//out, _ := json.Marshal(preout)
		//fmt.Fprintln(w, string(out))
	})

	//Topology data api
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		preout := make(map[string]string)
		var timeStart int64
		var timeStartStr string
		for _, v := range g.Cfg.Targets {
			if v.Addr != g.Cfg.Ip {
				timeStart = time.Now().Unix() - int64(v.Thdchecksec)
				timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
				preout[v.Name] = "false"
				g.DLock.Lock()
				sql := "SELECT ifnull(max(avgdelay),0) maxavgdelay, ifnull(max(losspk),0) maxlosspk ,count(1) Cnt FROM  `pinglog-" + v.Addr + "` where lastcheck > '" + timeStartStr + "' and (cast(avgdelay as double) > " + strconv.Itoa(v.Thdavgdelay) + " or cast(losspk as double) > " + strconv.Itoa(v.Thdloss) + ") "
				rows, err := g.Db.Query(sql)
				seelog.Debug("[func:/api/topology.json] ", sql)
				if err == nil {
					for rows.Next() {
						l := new(g.TopoLog)
						err := rows.Scan(&l.Maxavgdelay, &l.Maxlosspk, &l.Cnt)
						if err != nil {
							seelog.Error("[/api/topology.json] ", err)
						}
						sec, _ := strconv.Atoi(l.Cnt)
						if sec < v.Thdoccnum {
							preout[v.Name] = "true"
						}
					}
					rows.Close()
				} else {
					seelog.Error("[/api/topology.json] ", err)
				}
				g.DLock.Unlock()
			}
		}
		RenderJson(w, preout)
	})

	//alert api
	http.HandleFunc("/api/alert.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		var dtb string
		if len(r.Form["date"]) > 0 {
			dtb = r.Form["date"][0]
		} else {
			dateStartStr := time.Unix(time.Now().Unix(), 0).Format("20060102")
			dtb = "alertlog-" + dateStartStr
		}
		listpreout := []string{}
		g.DLock.Lock()
		sql := "SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'"
		lrows, lerr := g.Db.Query(sql)
		seelog.Debug("[func:/api/alert.json] ", sql)
		if lerr == nil {
			for lrows.Next() {
				var l string
				err := lrows.Scan(&l)
				if err != nil {
					seelog.Error("[/api/alert.json] ", err)
				}
				listpreout = append(listpreout, l)
			}
			lrows.Close()
		} else {
			seelog.Error("[/api/alert.json] ", lerr)
		}
		datapreout := []g.Alterdata{}
		rows, err := g.Db.Query("SELECT * FROM [" + dtb + "] where 1=1")
		if err == nil {
			for rows.Next() {
				l := new(g.Alterdata)
				err := rows.Scan(&l.Logtime, &l.Fromname, &l.Toname, &l.Alerttype)
				if err != nil {
					fmt.Println(err)
				}
				datapreout = append(datapreout, *l)
			}
			rows.Close()
		} else {
			seelog.Error("[/api/alert.json] ", err)
		}
		g.DLock.Unlock()
		lout, _ := json.Marshal(listpreout)
		dout, _ := json.Marshal(datapreout)
		fmt.Fprintln(w, "["+string(lout)+","+string(dout)+"]")
	})

	//save config
	http.HandleFunc("/api/saveconfig.json", func(w http.ResponseWriter, r *http.Request) {
		preout := make(map[string]string)
		r.ParseForm()
		preout["status"] = "false"
		if len(r.Form["password"]) == 0 {
			preout["info"] = "password empty!"
			RenderJson(w, preout)
			return
		}
		if r.Form["password"][0] != g.Cfg.Password {
			preout["info"] = "password error!"
			RenderJson(w, preout)
			return
		}
		if len(r.Form["config"]) == 0 {
			preout["info"] = "param error!"
			RenderJson(w, preout)
			return
		}
		//Base
		nconfig := g.Config{}
		nconfig.Targets = []g.Target{}
		err := json.Unmarshal([]byte(r.Form["config"][0]), &nconfig)
		if err != nil {
			preout["info"] = "decode json error!" + err.Error()
			RenderJson(w, preout)
			return
		}
		if nconfig.Name == "" {
			preout["info"] = "Agent Name illegal!"
			RenderJson(w, preout)
			return
		}
		if !ValidIP4(nconfig.Ip) {
			preout["info"] = "Agent Ip illegal!"
			RenderJson(w, preout)
			return
		}
		if nconfig.Timeout <= "0" {
			preout["info"] = "Timeout illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		//Alert
		if nconfig.Thdchecksec < 0 {
			preout["info"] = "Check Period illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Thdloss < 0 || nconfig.Thdloss > 100 {
			preout["info"] = "Loss Percent illegal!(<=0 and <=100)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Thdavgdelay <= 0 {
			preout["info"] = "Average Delay illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Thdoccnum < 0 {
			preout["info"] = "Occur Times illegal!(>=0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Alerthistory <= 0 {
			preout["info"] = "Archive Days !(>0)"
			RenderJson(w, preout)
			return
		}
		//Topology
		if nconfig.Tline <= "0" {
			preout["info"] = "Line Thickness  illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Tsymbolsize < "0" {
			preout["info"] = "Symbol Size illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Alertcycle <= 0 {
			preout["info"] = "Refresh illegal!(>0)"
			RenderJson(w, preout)
			return
		}
		//SmartPing NetWork
		targetcheck := true
		reminList := map[string]bool{}
		for _, v := range nconfig.Targets {
			if v.Name == "" {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(Empty Name Agent) "
				break
			}
			if !ValidIP4(v.Addr) {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal Addr Agent) "
				break
			}
			if v.Type != "CS" && v.Type != "C" {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal Type Agent) "
				break
			}
			if v.Thdchecksec <= 0 {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT CP Agent) "
				break
			}
			if v.Thdloss < 0 || v.Thdloss > 100 {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT LP Agent) "
				break
			}
			if v.Thdavgdelay <= 0 {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT AD Agent) "
				break
			}
			if v.Thdoccnum < 0 {
				targetcheck = false
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT OT Agent) "
				break
			}
			reminList["pinglog-"+v.Addr] = true
		}
		if !targetcheck {
			RenderJson(w, preout)
			return
		}
		preout["status"] = "true"
		nconfig.Db = g.Cfg.Db
		nconfig.Ver = g.Cfg.Ver
		nconfig.Port = g.Cfg.Port
		nconfig.Password = g.Cfg.Password
		g.Cfg = nconfig
		rrs, _ := json.Marshal(nconfig)
		var out bytes.Buffer
		errjson := json.Indent(&out, rrs, "", "\t")
		if errjson == nil {
			ioutil.WriteFile(g.GetRoot()+"/conf/"+"config.json", []byte(out.String()), 0644)
		} else {
			seelog.Error("[/api/saveconfig.json] ", err)
		}
		sql := ""
		listpreout := []string{}
		lrows, lerr := g.Db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%pinglog%'")
		if lerr == nil {
			for lrows.Next() {
				var l string
				err := lrows.Scan(&l)
				if err != nil {
					seelog.Error("[/api/saveconfig.json] ", err)
				}
				listpreout = append(listpreout, l)
			}
			lrows.Close()
		} else {
			seelog.Error("[/api/saveconfig.json] ", lerr)
		}
		for _, v := range listpreout {
			if _, ok := reminList[v]; !ok {
				sql = sql + "DROP TABLE [" + v + "];"
			}
		}
		funcs.SqlExec(sql)
		RenderJson(w, preout)

	})
}
