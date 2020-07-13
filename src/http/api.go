package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/funcs"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/nettools"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

func configApiRoutes() {

	//配置文件API
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		r.ParseForm()
		nconf := g.Config{}
		cfgJson, _ := json.Marshal(g.Cfg)
		json.Unmarshal(cfgJson, &nconf)
		nconf.Password = ""
		if !AuthAgentIp(r.RemoteAddr, false) {
			if nconf.Alert["SendEmailPassword"] != "" {
				nconf.Alert["SendEmailPassword"] = "samepasswordasbefore"
			}
		}
		//fmt.Print(g.Cfg.Alert["SendEmailPassword"])
		onconf, _ := json.Marshal(nconf)
		var out bytes.Buffer
		json.Indent(&out, onconf, "", "\t")
		o := out.String()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, o)
	})

	//Ping数据API
	http.HandleFunc("/api/ping.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
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
		querySql := "SELECT logtime,maxdelay,mindelay,avgdelay,losspk FROM `pinglog` where target='" + tableip + "' and logtime between '" + timeStartStr + "' and '" + timeEndStr + "' "
		rows, err := g.Db.Query(querySql)
		seelog.Debug("[func:/api/ping.json] Query ", querySql)
		if err != nil {
			seelog.Error("[func:/api/ping.json] Query ", err)
		} else {
			for rows.Next() {
				l := new(g.PingLog)
				err := rows.Scan(&l.Logtime, &l.Maxdelay, &l.Mindelay, &l.Avgdelay, &l.Losspk)
				if err != nil {
					seelog.Error("[/api/ping.json] Rows", err)
					continue
				}
				for n, v := range lastcheck {
					if v == l.Logtime {
						maxdelay[n] = l.Maxdelay
						mindelay[n] = l.Mindelay
						avgdelay[n] = l.Avgdelay
						losspk[n] = l.Losspk
						break
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
			"losspk":    losspk,
		}
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, preout)
	})

	//Ping拓扑API
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		preout := make(map[string]string)
		for _, v := range g.SelfCfg.Topology {
			if funcs.CheckAlertStatus(v) {
				preout[v["Addr"]] = "true"
			} else {
				preout[v["Addr"]] = "false"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, preout)
	})

	//报警API
	http.HandleFunc("/api/alert.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		type DateList struct {
			Ldate string
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		r.ParseForm()
		dtb := time.Unix(time.Now().Unix(), 0).Format("2006-01-02")
		if len(r.Form["date"]) > 0 {
			dtb = strings.Replace(r.Form["date"][0], "alertlog-", "", -1)
		}
		listpreout := []string{}
		datapreout := []g.AlertLog{}
		querySql := "select date(logtime) as ldate from alertlog group by date(logtime) order by logtime desc"
		rows, err := g.Db.Query(querySql)
		seelog.Debug("[func:/api/alert.json] Query ", querySql)
		if err != nil {
			seelog.Error("[func:/api/alert.json] Query ", err)
		} else {
			for rows.Next() {
				l := new(DateList)
				err := rows.Scan(&l.Ldate)
				if err != nil {
					seelog.Error("[/api/alert.json] Rows", err)
					continue
				}
				listpreout = append(listpreout, l.Ldate)
			}
			rows.Close()
		}
		querySql = "select logtime,targetname,targetip,tracert from alertlog where logtime between '" + dtb + " 00:00:00' and '" + dtb + " 23:59:59'"
		rows, err = g.Db.Query(querySql)
		seelog.Debug("[func:/api/alert.json] Query ", querySql)
		if err != nil {
			seelog.Error("[func:/api/alert.json] Query ", err)
		} else {
			for rows.Next() {
				l := new(g.AlertLog)
				err := rows.Scan(&l.Logtime, &l.Targetname, &l.Targetip, &l.Tracert)
				l.Fromname = g.Cfg.Name
				l.Fromip = g.Cfg.Addr
				if err != nil {
					seelog.Error("[/api/alert.json] Rows", err)
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

	//全国延迟API
	http.HandleFunc("/api/mapping.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
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
		type Mapjson struct {
			Mapjson string
		}
		chinaMp := g.ChinaMp{}
		chinaMp.Text = g.Cfg.Name
		chinaMp.Subtext = dataKey
		chinaMp.Avgdelay = map[string][]g.MapVal{}
		chinaMp.Avgdelay["ctcc"] = []g.MapVal{}
		chinaMp.Avgdelay["cucc"] = []g.MapVal{}
		chinaMp.Avgdelay["cmcc"] = []g.MapVal{}
		g.DLock.Lock()
		querySql := "select mapjson from mappinglog where logtime = '" + dataKey + "'"
		rows, err := g.Db.Query(querySql)
		g.DLock.Unlock()
		seelog.Debug("[func:/api/mapping.json] Query ", querySql)
		if err != nil {
			seelog.Error("[func:/api/mapping.json] Query ", err)
		} else {
			for rows.Next() {
				l := new(Mapjson)
				err := rows.Scan(&l.Mapjson)
				if err != nil {
					seelog.Error("[/api/mapping.json] Rows", err)
					continue
				}
				json.Unmarshal([]byte(l.Mapjson), &chinaMp.Avgdelay)
			}
			rows.Close()
		}
		w.Header().Set("Content-Type", "application/json")
		RenderJson(w, chinaMp)
	})

	//检测工具API
	http.HandleFunc("/api/tools.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
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

	//保存配置文件
	http.HandleFunc("/api/saveconfig.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		preout := make(map[string]string)
		r.ParseForm()
		preout["status"] = "false"
		if len(r.Form["password"]) == 0 || r.Form["password"][0] != g.Cfg.Password {
			preout["info"] = "密码错误!"
			RenderJson(w, preout)
			return
		}
		if len(r.Form["config"]) == 0 {
			preout["info"] = "参数错误!"
			RenderJson(w, preout)
			return
		}
		nconfig := g.Config{}
		err := json.Unmarshal([]byte(r.Form["config"][0]), &nconfig)
		if err != nil {
			preout["info"] = "配置文件解析错误!" + err.Error()
			RenderJson(w, preout)
			return
		}
		if nconfig.Name == "" {
			preout["info"] = "本机节点名称为空!"
			RenderJson(w, preout)
			return
		}
		if !ValidIP4(nconfig.Addr) {
			preout["info"] = "非法本机节点IP!"
			RenderJson(w, preout)
			return
		}
		//Base
		if _, ok := nconfig.Base["Timeout"]; !ok || nconfig.Base["Timeout"] <= 0 {
			preout["info"] = "非法超时时间!(>0)"
			RenderJson(w, preout)
			return
		}
		if _, ok := nconfig.Base["Archive"]; !ok || nconfig.Base["Archive"] <= 0 {
			preout["info"] = "非法存档天数!(>0)"
			RenderJson(w, preout)
			return
		}
		if _, ok := nconfig.Base["Refresh"]; !ok || nconfig.Base["Refresh"] <= 0 {
			preout["info"] = "非法刷新频率!(>0)"
			RenderJson(w, preout)
			return
		}
		//Topology
		if _, ok := nconfig.Topology["Tline"]; !ok || nconfig.Topology["Tline"] <= "0" {
			preout["info"] = "非法拓扑连线粗细(>0)"
			RenderJson(w, preout)
			return
		}
		if _, ok := nconfig.Topology["Tsymbolsize"]; !ok || nconfig.Topology["Tsymbolsize"] <= "0" {
			preout["info"] = "非法拓扑形状大小!(>0)"
			RenderJson(w, preout)
			return
		}
		if nconfig.Toollimit < 0 {
			preout["info"] = "非法检测工具限定频率!(>=0)"
			RenderJson(w, preout)
			return
		}
		//Network
		for k, network := range nconfig.Network {
			if !ValidIP4(network.Addr) || !ValidIP4(k) {
				preout["info"] = "Ping节点测试网络信息错误!(非法节点IP地址 " + k + ")"
				RenderJson(w, preout)
				return
			}
			if network.Name == "" {
				preout["info"] = "Ping节点测试网络信息错误!( " + k + " 节点名称为空)"
				RenderJson(w, preout)
				return
			}
			for _, topology := range network.Topology {
				if _, ok := topology["Thdchecksec"]; !ok {
					preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，秒) "
					RenderJson(w, preout)
					return
				} else {
					Thdchecksec, err := strconv.Atoi(topology["Thdchecksec"])
					if err != nil || Thdchecksec <= 0 {
						preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，>0 秒  ) "
						RenderJson(w, preout)
						return
					}
				}
				if _, ok := topology["Thdloss"]; !ok {
					preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，%) "
					RenderJson(w, preout)
					return
				} else {
					Thdloss, err := strconv.Atoi(topology["Thdloss"])
					if err != nil || (Thdloss < 0 || Thdloss > 100) {
						preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，0 <= % <=100  ) "
						RenderJson(w, preout)
						return
					}
				}
				if _, ok := topology["Thdavgdelay"]; !ok {
					preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，ms) "
					RenderJson(w, preout)
					return
				} else {
					Thdavgdelay, err := strconv.Atoi(topology["Thdavgdelay"])
					if err != nil || Thdavgdelay <= 0 {
						preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，> 0 ms  ) "
						RenderJson(w, preout)
						return
					}
				}
				if _, ok := topology["Thdoccnum"]; !ok {
					preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，次) "
					RenderJson(w, preout)
					return
				} else {
					Thdoccnum, err := strconv.Atoi(topology["Thdoccnum"])
					if err != nil || Thdoccnum <= 0 {
						preout["info"] = "Ping节点测试网络信息错误!( " + k + "->" + topology["Addr"] + " 非法拓扑报警规则，> 0 次  ) "
						RenderJson(w, preout)
						return
					}
				}
			}
		}
		//ChinaMap
		for _, provVal := range nconfig.Chinamap {
			for _, telcomVal := range provVal {
				for _, ip := range telcomVal {
					if ip != "" && !ValidIP4(ip) {
						preout["info"] = "Mapping Ip illegal!"
						RenderJson(w, preout)
						return
					}
				}
			}
		}
		nconfig.Ver = g.Cfg.Ver
		nconfig.Port = g.Cfg.Port
		nconfig.Password = g.Cfg.Password
		if nconfig.Alert["SendEmailPassword"] == "samepasswordasbefore" {
			nconfig.Alert["SendEmailPassword"] = g.Cfg.Alert["SendEmailPassword"]
		}
		g.Cfg = nconfig
		g.SelfCfg = g.Cfg.Network[g.Cfg.Addr]
		saveerr := g.SaveConfig()
		if saveerr != nil {
			preout["info"] = saveerr.Error()
			RenderJson(w, preout)
			return
		}
		preout["status"] = "true"
		RenderJson(w, preout)
	})

	//发送测试邮件
	http.HandleFunc("/api/sendmailtest.json", func(w http.ResponseWriter, r *http.Request) {
		if !AuthUserIp(r.RemoteAddr) && !AuthAgentIp(r.RemoteAddr, true) {
			o := "Your ip address (" + r.RemoteAddr + ")  is not allowed to access this site!"
			http.Error(w, o, 401)
			return
		}
		preout := make(map[string]string)
		r.ParseForm()
		preout["status"] = "false"
		if len(r.Form["EmailHost"]) == 0 {
			preout["info"] = "邮件服务器不能为空!"
			RenderJson(w, preout)
			return
		}
		if len(r.Form["SendEmailAccount"]) == 0 {
			preout["info"] = "发件邮件不能为空!"
			RenderJson(w, preout)
			return
		}
		if len(r.Form["SendEmailPassword"]) == 0 {
			preout["info"] = "发件邮箱密码不能为空!"
			RenderJson(w, preout)
			return
		}
		if len(r.Form["RevcEmailList"]) == 0 {
			preout["info"] = "收件邮箱列表不能为空!"
			RenderJson(w, preout)
			return
		}

		err := funcs.SendMail(r.Form["SendEmailAccount"][0], r.Form["SendEmailPassword"][0], r.Form["EmailHost"][0], r.Form["RevcEmailList"][0], "报警测试邮件 - SmartPing", "报警测试邮件")
		if err != nil {
			preout["info"] = err.Error()
			RenderJson(w, preout)
			return
		}
		preout["status"] = "true"
		RenderJson(w, preout)
	})

	//Ping画图
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
		timeout := time.Duration(time.Duration(g.Cfg.Base["Timeout"]) * time.Second)
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
				//NameStyle: chart.StyleShow(),
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

	//代理访问
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
		to := strconv.Itoa(g.Cfg.Base["Timeout"])
		if len(r.Form["t"]) > 0 {
			to = r.Form["t"][0]
		}
		url := strings.Replace(strings.Replace(r.Form["g"][0], "%26", "&", -1), " ", "%20", -1)
		defaultto, err := strconv.Atoi(to)
		if err != nil {
			o := "Timeout Param Error!"
			http.Error(w, o, 406)
			return
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
