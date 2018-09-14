package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gy-games-libs/seelog"
	"github.com/gy-games/smartping/src/g"
	"net/http"
	"strconv"
	"time"
)

func configApiRoutes() {

	//config api
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		nconf := g.Config{}
		if len(r.Form["cloudendpoint"]) > 0 {
			cloudnconf, err := g.SaveCloudConfig(r.Form["cloudendpoint"][0], false)
			if err != nil {
				preout := make(map[string]string)
				preout["status"] = "false"
				preout["info"] = cloudnconf.Name+"("+err.Error()+")"
				RenderJson(w, preout)
				return
			}
			nconf = cloudnconf
		} else {
			nconf = g.Cfg
		}
		nconf.Password = ""
		onconf, _ := json.Marshal(nconf)
		var out bytes.Buffer
		json.Indent(&out, onconf, "", "\t")
		o := out.String()
		fmt.Fprintln(w, o)
	})

	//graph data api
	http.HandleFunc("/api/ping.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
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
		querySql := "SELECT logtime,maxdelay,mindelay,avgdelay,sendpk,revcpk,losspk,lastcheck FROM `pinglog-" + tableip + "` where 1=1 and lastcheck between '" + timeStartStr + "' and '" + timeEndStr + "' " + where + ""
		rows, err := g.Db.Query(querySql)
		g.DLock.Unlock()
		seelog.Debug("[func:/api/ping.json] Query ", querySql)
		if err != nil {
			seelog.Error("[func:/api/ping.json] Query ", err)
		} else {
			for rows.Next() {
				l := new(g.LogInfo)
				err := rows.Scan(&l.Logtime, &l.Maxdelay, &l.Mindelay, &l.Avgdelay, &l.Sendpk, &l.Revcpk, &l.Losspk, &l.Lastcheck)
				if err != nil {
					seelog.Error("[/api/ping.json] Rows ", err)
					continue
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
	})

	//Topology data api
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
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
				querySql := "SELECT ifnull(max(avgdelay),0) maxavgdelay, ifnull(max(losspk),0) maxlosspk ,count(1) Cnt FROM  `pinglog-" + v.Addr + "` where lastcheck > '" + timeStartStr + "' and (cast(avgdelay as double) > " + strconv.Itoa(v.Thdavgdelay) + " or cast(losspk as double) > " + strconv.Itoa(v.Thdloss) + ") "
				rows, err := g.Db.Query(querySql)
				g.DLock.Unlock()
				seelog.Debug("[func:/api/topology.json] Query Topology", querySql)
				if err != nil {
					seelog.Error("[/api/topology.json] ", err)
				} else {
					for rows.Next() {
						l := new(g.TopoLog)
						err := rows.Scan(&l.Maxavgdelay, &l.Maxlosspk, &l.Cnt)
						if err != nil {
							seelog.Error("[/api/topology.json] ", err)
							preout[v.Name] = "unknown"
							continue
						}
						sec, _ := strconv.Atoi(l.Cnt)
						if sec < v.Thdoccnum {
							preout[v.Name] = "true"
						}
					}
					rows.Close()
				}
			}
		}
		RenderJson(w, preout)
	})

	//alert api
	http.HandleFunc("/api/alert.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
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
		querySql := "SELECT name FROM [sqlite_master] where type='table' and name like '%alertlog%'"
		lrows, lerr := g.Db.Query(querySql)
		g.DLock.Unlock()
		seelog.Debug("[func:/api/alert.json] Query Table List", querySql)
		if lerr != nil {
			seelog.Error("[/api/alert.json] ", lerr)
		} else {
			for lrows.Next() {
				var l string
				err := lrows.Scan(&l)
				if err != nil {
					seelog.Error("[/api/alert.json]  Rows Table List", err)
					continue
				}
				listpreout = append(listpreout, l)
			}
			lrows.Close()
		}
		datapreout := []g.Alterdata{}
		g.DLock.Lock()
		querySql = "SELECT * FROM [" + dtb + "] where 1=1"
		rows, err := g.Db.Query(querySql)
		g.DLock.Unlock()
		seelog.Debug("[func:/api/alert.json] Query Detail Data", querySql)
		if err != nil {
			seelog.Error("[/api/alert.json] Query Detail Data", err)
		} else {
			for rows.Next() {
				l := new(g.Alterdata)
				err := rows.Scan(&l.Logtime, &l.Fromname, &l.Toname, &l.Tracert)
				if err != nil {
					seelog.Error("[/api/alert.json]  Rows Detail Data", err)
					continue
				}
				datapreout = append(datapreout, *l)
			}
			rows.Close()
		}

		lout, _ := json.Marshal(listpreout)
		dout, _ := json.Marshal(datapreout)
		fmt.Fprintln(w, "["+string(lout)+","+string(dout)+"]")
	})

	//save config
	http.HandleFunc("/api/saveconfig.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
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
		for _, v := range nconfig.Targets {
			if v.Name == "" {
				preout["info"] = "SmartPing Network Info illegal!(Empty Name Agent) "
				RenderJson(w, preout)
				return
			}
			if !ValidIP4(v.Addr) {
				preout["info"] = "SmartPing Network Info illegal!(illegal Addr Agent) "
				RenderJson(w, preout)
				return
			}
			if v.Type != "CS" && v.Type != "C" {
				preout["info"] = "SmartPing Network Info illegal!(illegal Type Agent) "
				RenderJson(w, preout)
				return
			}
			if v.Thdchecksec <= 0 {
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT CP Agent) "
				RenderJson(w, preout)
				return
			}
			if v.Thdloss < 0 || v.Thdloss > 100 {
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT LP Agent) "
				RenderJson(w, preout)
				return
			}
			if v.Thdavgdelay <= 0 {
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT AD Agent) "
				RenderJson(w, preout)
				return
			}
			if v.Thdoccnum < 0 {
				preout["info"] = "SmartPing Network Info illegal!(illegal ALERT OT Agent) "
				RenderJson(w, preout)
				return
			}
		}
		//nconfig.Db = g.Cfg.Db
		nconfig.Ver = g.Cfg.Ver
		//nconfig.Port = g.Cfg.Port
		nconfig.Mode = "local"
		nconfig.Password = g.Cfg.Password
		nconfig.Port = g.Cfg.Port
		g.Cfg = nconfig
		saveerr := g.SaveConfig()
		if saveerr != nil {
			preout["info"] = saveerr.Error()
			RenderJson(w, preout)
			return
		}
		preout["status"] = "true"
		RenderJson(w, preout)
	})

	//save cloud config
	http.HandleFunc("/api/savecloudconfig.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
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
		nconfig := g.Config{}
		nconfig.Targets = []g.Target{}
		err := json.Unmarshal([]byte(r.Form["config"][0]), &nconfig)
		if err != nil {
			preout["info"] = "Decode json error!" + err.Error()
			RenderJson(w, preout)
			return
		}
		if nconfig.Name == "" {
			preout["info"] = "Agent Name illegal!"
			RenderJson(w, preout)
			return
		}
		if nconfig.Cendpoint == "" {
			preout["info"] = "Cloud Endpoint illegal!"
			RenderJson(w, preout)
			return
		}
		if !ValidIP4(nconfig.Ip) {
			preout["info"] = "Agent Ip illegal!"
			RenderJson(w, preout)
			return
		}
		_, err = g.SaveCloudConfig(nconfig.Cendpoint, true)
		if err != nil {
			preout["info"] = err.Error()
			RenderJson(w, preout)
			return
		}
		g.Cfg.Name = nconfig.Name
		g.Cfg.Cendpoint = nconfig.Cendpoint
		g.Cfg.Alertsound = nconfig.Alertsound
		g.Cfg.Ip = nconfig.Ip
		g.Cfg.Password = g.Cfg.Password
		g.Cfg.Cstatus = true
		saveerr := g.SaveConfig()
		if saveerr != nil {
			preout["info"] = saveerr.Error()
			RenderJson(w, preout)
			return
		}
		preout["status"] = "true"
		RenderJson(w, preout)
	})
}
