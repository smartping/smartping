package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/smartping/smartping/src/g"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"github.com/boltdb/bolt"
	"github.com/cihub/seelog"
	//"strings"
	"github.com/smartping/smartping/src/funcs"
	//"context"
	"strings"
)

func configApiRoutes() {

	//config api
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		r.ParseForm()
		nconf := g.Config{}
		if len(r.Form["cloudendpoint"]) > 0 {
			cloudnconf, err := g.SaveCloudConfig(r.Form["cloudendpoint"][0], false)
			if err != nil {
				preout := make(map[string]string)
				preout["status"] = "false"
				preout["info"] = cloudnconf.Name + "(" + err.Error() + ")"
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
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, o)
	})

	//graph data api
	http.HandleFunc("/api/ping.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		r.ParseForm()
		if len(r.Form["ip"]) == 0 {
			o := "Missing Param !"
			http.Error(w, o, 406)
			return
		}
		var tableip string
		var timeStart int64
		var timeEnd int64
		var timeStartStr string
		var timeEndStr string
		tableip = r.Form["ip"][0]
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
		//var sendpk []string
		//var revcpk []string
		var losspk []string
		timwwnum := map[string]int{}
		for i := 0; i < cnt+1; i++ {
			ntime:=time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			timwwnum[ntime]=i
			lastcheck = append(lastcheck, ntime)
			maxdelay = append(maxdelay, "0")
			mindelay = append(mindelay, "0")
			avgdelay = append(avgdelay, "0")
			//sendpk = append(sendpk, "0")
			//revcpk = append(revcpk, "0")
			losspk = append(losspk, "0")
			timeStart = timeStart + 60
		}
		db:=g.GetDb("ping",tableip)
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("pinglog"))
			if b==nil{
				return nil
			}
			c:= b.Cursor()
			min := []byte(timeStartStr[8:])
			max := []byte(timeEndStr[8:])
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				l := new(g.PingLog)
				err := json.Unmarshal(v,&l)
				if err!=nil{
					continue
				}
				if l.Logtime >= timeStartStr && l.Logtime <= timeEndStr{
					maxdelay[timwwnum[l.Logtime]] = l.Maxdelay
					mindelay[timwwnum[l.Logtime]] = l.Mindelay
					avgdelay[timwwnum[l.Logtime]] = l.Avgdelay
					losspk[timwwnum[l.Logtime]] = l.Losspk
				}
			}
			return nil
		})
		preout := map[string][]string{
			"lastcheck": lastcheck,
			"maxdelay":  maxdelay,
			"mindelay":  mindelay,
			"avgdelay":  avgdelay,
			"losspk":    losspk,
		}
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, preout)
	})

	//topology data api
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		preout := make(map[string]string)
		for _, v := range g.Cfg.Targets {
			if v.Addr != g.Cfg.Ip {
				preout[v.Name] =  funcs.CheckAlertStatus(v)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, preout)

	})

	//alert api
	http.HandleFunc("/api/alert.json", func(w http.ResponseWriter, r *http.Request) {

		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		var dtb string
		if len(r.Form["date"]) > 0 {
			dtb = strings.Replace(r.Form["date"][0],"alertlog-","",-1)

		} else {
			dtb = time.Unix(time.Now().Unix(), 0).Format("20060102")
		}
		listpreout := []string{}
		datapreout := []g.AlertLog{}
		db:=g.GetDb("alert",dtb)
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("alertlog"))
			if b==nil{
				return nil
			}
			b.ForEach(func(k, v []byte) error {
				l := g.AlertLog{}
				err:= json.Unmarshal(v,&l)
				if err==nil{
					datapreout = append(datapreout, l)
				}
				return nil
			})
			return nil
		})
		lout, _ := json.Marshal(listpreout)
		dout, _ := json.Marshal(datapreout)
		fmt.Fprintln(w, "["+string(lout)+","+string(dout)+"]")
	})

	//save config
	http.HandleFunc("/api/saveconfig.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
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
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
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

	//show graph
	http.HandleFunc("/api/graph.png", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		r.ParseForm()
		if len(r.Form["g"]) == 0 {
			GraphText(83, 70, "GET PARAM ERROR").Save(w)
			return
		}
		url := r.Form["g"][0]
		config := g.PingStMini{}
		defaultto,err := strconv.Atoi(g.Cfg.Timeout)
		if err!=nil{
			defaultto = 3
		}
		timeout := time.Duration(time.Duration(defaultto) * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(url)
		if err != nil {
			GraphText(80, 70, "REQUEST API ERROR").Save(w)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == 401 {
			GraphText(80, 70, "401-UNAUTHORIZED").Save(w)
			return
		}
		if resp.StatusCode != 200 {
			GraphText(85, 70, "ERROR CODE "+strconv.Itoa(resp.StatusCode)).Save(w)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &config)
		if err != nil {
			GraphText(80, 70, "PARSE DATA ERROR").Save(w)
			return
		}
		Xals := []float64{}
		AvgDelay := []float64{}
		LossPk := []float64{}
		Bkg := []float64{}
		MaxDelay := 0.0
		for i := 0; i < len(config.LossPk); i = i + 1 {
			avg, _ := strconv.ParseFloat(config.AvgDelay[i], 64)
			if MaxDelay < avg {
				MaxDelay = avg
			}
			AvgDelay = append(AvgDelay, avg)
			losspk, _ := strconv.ParseFloat(config.LossPk[i], 64)
			LossPk = append(LossPk, losspk)
			Xals = append(Xals, float64(i))
			Bkg = append(Bkg, 100.0)
		}
		graph := chart.Chart{
			Width:  300 * 3,
			Height: 130 * 3,
			Background: chart.Style{
				FillColor: drawing.Color{249, 246, 241, 255},
			},
			XAxis: chart.XAxis{
				Style: chart.Style{
					Show:     true,
					FontSize: 20,
				},
				TickPosition: chart.TickPositionBetweenTicks,
				ValueFormatter: func(v interface{}) string {
					return config.Lastcheck[int(v.(float64))][11:16]
				},
			},
			YAxis: chart.YAxis{
				Style: chart.Style{
					Show:     true,
					FontSize: 20,
				},
				Range: &chart.ContinuousRange{
					Min: 0.0,
					Max: 100.0,
				},
				ValueFormatter: func(v interface{}) string {
					if vf, isFloat := v.(float64); isFloat {
						return fmt.Sprintf("%0.0f", vf)
					}
					return ""
				},
			},
			YAxisSecondary: chart.YAxis{
				NameStyle: chart.StyleShow(),
				Style: chart.Style{
					Show:     true,
					FontSize: 20,
				},
				Range: &chart.ContinuousRange{
					Min: 0.0,
					Max: MaxDelay + MaxDelay/10,
				},
				ValueFormatter: func(v interface{}) string {
					if vf, isFloat := v.(float64); isFloat {
						return fmt.Sprintf("%0.0f", vf)
					}
					return ""
				},
			},
			Series: []chart.Series{
				chart.ContinuousSeries{
					Style: chart.Style{
						Show:        true,
						StrokeColor: drawing.Color{249, 246, 241, 255},
						FillColor:   drawing.Color{249, 246, 241, 255},
					},
					XValues: Xals,
					YValues: Bkg,
				},
				chart.ContinuousSeries{
					Style: chart.Style{
						Show:        true,
						StrokeColor: drawing.Color{0, 204, 102, 200},
						FillColor:   drawing.Color{0, 204, 102, 200},
					},
					XValues: Xals,
					YValues: AvgDelay,
					YAxis:   chart.YAxisSecondary,
				},
				chart.ContinuousSeries{
					Style: chart.Style{
						Show:        true,
						StrokeColor: drawing.Color{255, 0, 0, 200},
						FillColor:   drawing.Color{255, 0, 0, 200},
					},
					XValues: Xals,
					YValues: LossPk,
				},
			},
		}
		graph.Render(chart.PNG, w)

	})

	//remote apip roxy
	http.HandleFunc("/api/agentproxy.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		//w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		if len(r.Form["g"]) == 0 {
			o := "Param Error!"
			http.Error(w, o, 406)
		}
		url := strings.Replace(strings.Replace(r.Form["g"][0],"%26","&",-1)," ","%20",-1)
		seelog.Debug("[/api/agentproxy.json] GET ",url)
		defaultto,err := strconv.Atoi(g.Cfg.Timeout)
		if err!=nil{
			defaultto = 3
		}
		timeout := time.Duration(time.Duration(defaultto) * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		resp, err := client.Get(url)
		if err != nil {
			o := "Request Remote Data Error:" + err.Error()
			http.Error(w, o, 503)
			return
		}
		defer resp.Body.Close()
		resCode := resp.StatusCode
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			o := "Read Remote Data Error:" + err.Error()
			http.Error(w, o, 503)
			return
		}
		if resCode != 200 {
			o := "Get Remote Data Status Error"
			http.Error(w, o, resCode)
		}
		var out bytes.Buffer
		json.Indent(&out, body, "", "\t")
		o := out.String()
		fmt.Fprintln(w, o)
		//RenderJson(w, body)
	})

}
