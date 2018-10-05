package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/funcs"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
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
			ntime := time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			timwwnum[ntime] = i
			lastcheck = append(lastcheck, ntime)
			maxdelay = append(maxdelay, "0")
			mindelay = append(mindelay, "0")
			avgdelay = append(avgdelay, "0")
			losspk = append(losspk, "0")
			timeStart = timeStart + 60
		}
		db := g.GetDb("ping", tableip)
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("pinglog"))
			if b == nil {
				return nil
			}
			c := b.Cursor()
			min := []byte(timeStartStr[8:])
			max := []byte(timeEndStr[8:])
			for k, v := c.Seek(min); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
				l := new(g.PingLog)
				err := json.Unmarshal(v, &l)
				if err != nil {
					continue
				}
				if l.Logtime >= timeStartStr && l.Logtime <= timeEndStr {
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
			if v.Addr != g.Cfg.Ip && v.Topo {
				preout[v.Name] = funcs.CheckAlertStatus(v)
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
			dtb = strings.Replace(r.Form["date"][0], "alertlog-", "", -1)

		} else {
			dtb = time.Unix(time.Now().Unix(), 0).Format("20060102")
		}
		listpreout := []string{}
		datapreout := []g.AlertLog{}
		db := g.GetDb("alert", dtb)
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("alertlog"))
			if b == nil {
				return nil
			}
			b.ForEach(func(k, v []byte) error {
				l := g.AlertLog{}
				err := json.Unmarshal(v, &l)
				if err == nil {
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

	//mapping
	http.HandleFunc("/api/mapping.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		m, _ := time.ParseDuration("-1m")
		dataKey := time.Now().Add(m).Format("2006-01-02 15:04")
		r.ParseForm()
		if len(r.Form["d"]) > 0 {
			dataKey = r.Form["d"][0]
		}
		chinaMp := g.ChinaMp{}
		chinaMp.Text = "No Data"
		chinaMp.Subtext = dataKey
		chinaMp.Avgdelay = map[string][]g.MapVal{}
		chinaMp.Avgdelay["ctcc"] = []g.MapVal{}
		chinaMp.Avgdelay["cucc"] = []g.MapVal{}
		chinaMp.Avgdelay["cmcc"] = []g.MapVal{}
		bucketName := time.Unix(time.Now().Unix(), 0).Format("20060102")
		db := g.GetDb("mapping", bucketName)
		db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("mapping"))
			if b == nil {
				return nil
			}
			v := b.Get([]byte(dataKey))
			if v != nil {
				json.Unmarshal(v, &chinaMp)
			}
			return nil
		})
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, chinaMp)
	})

	//tools
	http.HandleFunc("/api/tools.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		preout := g.ToolsRes{}
		preout.Status = "false"
		r.ParseForm()
		if len(r.Form["t"]) == 0 {
			preout.Error = "target empty!"
			RenderJson(w, preout)
			return
		}
		nowtime := int(time.Now().Unix())
		if _, ok := g.ToolLimit[r.RemoteAddr]; ok {
			if (nowtime - g.ToolLimit[r.RemoteAddr]) <= g.Cfg.Toollimit {
				preout.Error = "Time Limit Exceeded!"
				RenderJson(w, preout)
				return
			}
		}
		g.ToolLimit[r.RemoteAddr] = nowtime
		target := strings.Replace(strings.Replace(r.Form["t"][0], "https://", "", -1), "http://", "", -1)
		preout.Ping = g.PingSt{}
		preout.Ping.MinDelay = -1
		lossPK := 0
		ipaddr, err := net.ResolveIPAddr("ip", target)
		if err != nil {
			preout.Error = "Unable to resolve destination host"
			RenderJson(w, preout)
			return
		}
		preout.Ip = ipaddr.String()
		var channel chan float64 = make(chan float64, 5)
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				delay, err := nettools.RunPing(ipaddr, 3*time.Second, 64, i)
				if err != nil {
					channel <- -1.00
				} else {
					channel <- delay
				}
				wg.Done()
			}()
			seelog.Debug(i)
			time.Sleep(time.Duration(100 * time.Millisecond))
		}
		wg.Wait()
		for i := 0; i < 5; i++ {
			select {
			case delay := <-channel:
				if delay != -1.00 {
					preout.Ping.AvgDelay = preout.Ping.AvgDelay + delay
					if preout.Ping.MaxDelay < delay {
						preout.Ping.MaxDelay = delay
					}
					if preout.Ping.MinDelay == -1 || preout.Ping.MinDelay > delay {
						preout.Ping.MinDelay = delay
					}
					preout.Ping.RevcPk = preout.Ping.RevcPk + 1
				} else {
					lossPK = lossPK + 1
				}
				preout.Ping.SendPk = preout.Ping.SendPk + 1
				preout.Ping.LossPk = int((float64(lossPK) / float64(preout.Ping.SendPk)) * 100)
			}
		}
		if preout.Ping.RevcPk > 0 {
			preout.Ping.AvgDelay = preout.Ping.AvgDelay / float64(preout.Ping.RevcPk)
		} else {
			preout.Ping.AvgDelay = 3000
			preout.Ping.MinDelay = 3000
			preout.Ping.MaxDelay = 3000
		}
		preout.Status = "true"
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, preout)
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
		if nconfig.Archive <= 0 {
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
		if nconfig.Refresh <= 0 {
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

		//Map
		for _, telcomVal := range nconfig.Chinamap {
			for _, provVal := range telcomVal {
				for _, ip := range provVal {
					if ip != "" && !ValidIP4(ip) {
						preout["info"] = "Mapping Ip illegal!"
						RenderJson(w, preout)
						return
					}
				}
			}
		}
		seelog.Debug(nconfig)
		//return
		nconfig.Ver = g.Cfg.Ver
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
		if nconfig.Endpoint == "" {
			preout["info"] = "Cloud Endpoint illegal!"
			RenderJson(w, preout)
			return
		}
		if !ValidIP4(nconfig.Ip) {
			preout["info"] = "Agent Ip illegal!"
			RenderJson(w, preout)
			return
		}
		_, err = g.SaveCloudConfig(nconfig.Endpoint, true)
		if err != nil {
			preout["info"] = err.Error()
			RenderJson(w, preout)
			return
		}
		g.Cfg.Name = nconfig.Name
		g.Cfg.Endpoint = nconfig.Endpoint
		//g.Cfg.Tsound = nconfig.Tsound
		g.Cfg.Ip = nconfig.Ip
		g.Cfg.Password = g.Cfg.Password
		g.Cfg.Status = true
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
		defaultto, err := strconv.Atoi(g.Cfg.Timeout)
		if err != nil {
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
	http.HandleFunc("/api/proxy.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		r.ParseForm()
		if len(r.Form["g"]) == 0 {
			o := "Url Param Error!"
			http.Error(w, o, 406)
			return
		}
		to := g.Cfg.Timeout
		if len(r.Form["t"]) > 0 {
			to = r.Form["t"][0]
		}
		url := strings.Replace(strings.Replace(r.Form["g"][0], "%26", "&", -1), " ", "%20", -1)
		defaultto, err := strconv.Atoi(to)
		if err != nil {
			o := "Timeout Param Error!"
			http.Error(w, o, 406)
			return
			//defaultto = 3
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
	})

}
