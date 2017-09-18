package http

import (
	"../funcs"
	"../g"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

func configApiRoutes(db *sql.DB, config *g.Config) {

	//config api
	http.HandleFunc("/api/config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Type", "application/json")
		nconf := g.Config{}
		nconf = *config
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
		rows, err := db.Query("SELECT logtime,maxdelay,mindelay,avgdelay,sendpk,revcpk,losspk,lastcheck FROM `pinglog-" + tableip + "` where 1=1 and lastcheck between '" + timeStartStr + "' and '" + timeEndStr + "' " + where + "")
		if err == nil {
			for rows.Next() {
				l := new(g.LogInfo)
				err := rows.Scan(&l.Logtime, &l.Maxdelay, &l.Mindelay, &l.Avgdelay, &l.Sendpk, &l.Revcpk, &l.Losspk, &l.Lastcheck)
				if err != nil {
					log.Println(err)
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
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
	})

	//Topology data api
	http.HandleFunc("/api/topology.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		preout := make(map[string]string)
		var timeStart int64
		var timeStartStr string
		for _, v := range config.Targets {
			timeStart = time.Now().Unix() - int64(v.Thdchecksec)
			timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			preout[v.Name] = "false"
			g.DLock.Lock()
			rows, err := db.Query("SELECT max(avgdelay) maxavgdelay, max(losspk) maxlosspk ,count(1) Cnt FROM  `pinglog-" + v.Addr + "` where lastcheck > '" + timeStartStr + "' and (cast(avgdelay as double) > " + strconv.Itoa(v.Thdavgdelay) + " or cast(losspk as double) > " + strconv.Itoa(v.Thdloss) + ") ")
			if err == nil {
				for rows.Next() {
					l := new(g.TopoLog)
					rows.Scan(&l.Maxavgdelay, &l.Maxlosspk, &l.Cnt)
					sec, _ := strconv.Atoi(l.Cnt)
					if sec < v.Thdoccnum {
						preout[v.Name] = "true"
					}
				}
				rows.Close()
			}

			g.DLock.Unlock()
		}
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
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
		datapreout := []g.Alterdata{}
		rows, err := db.Query("SELECT * FROM [" + dtb + "] where 1=1")
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
		if len(r.Form["password"]) > 0 {
			if r.Form["password"][0] == config.Password {
				if len(r.Form["config"]) > 0 {
					nconfig := g.Config{}
					nconfig.Targets = []g.Target{}
					err := json.Unmarshal([]byte(r.Form["config"][0]), &nconfig)
					if err == nil {
						if nconfig.Name != "" {
							if funcs.ValidIP4(nconfig.Ip) {
								if nconfig.Timeout > "0" {
									if nconfig.Thdchecksec > 0 {
										if nconfig.Thdloss >= 0 && nconfig.Thdloss <= 100 {
											if nconfig.Thdavgdelay > 0 {
												if nconfig.Thdoccnum >= 0 {
													if nconfig.Alerthistory > 0 {
														if nconfig.Tline > "0" {
															if nconfig.Tsymbolsize > "0" {
																if nconfig.Alertcycle > 0 {
																	targetcheck := true
																	reminList := map[string]bool{}
																	for _, v := range nconfig.Targets {
																		if v.Name == "" {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(Empty Name Agent) "
																			break
																		}
																		if !funcs.ValidIP4(v.Addr) {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal Addr Agent) "
																			break
																		}
																		if v.Type != "CS" && v.Type != "C" {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal Type Agent) "
																			break
																		}
																		if v.Thdchecksec <= 0 {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal ALERT CP Agent) "
																			break
																		}
																		if v.Thdloss < 0 || v.Thdloss > 100 {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal ALERT LP Agent) "
																			break
																		}
																		if v.Thdavgdelay <= 0 {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal ALERT AD Agent) "
																			break
																		}
																		if v.Thdoccnum < 0 {
																			targetcheck = false
																			preout["info"] = "Agent List Info illegal!(illegal ALERT OT Agent) "
																			break
																		}
																		reminList["pinglog-"+v.Addr] = true
																	}
																	if targetcheck {
																		preout["status"] = "true"
																		nconfig.Db = config.Db
																		nconfig.Ver = config.Ver
																		nconfig.Port = config.Port
																		nconfig.Password = config.Password
																		*config = nconfig
																		r, _ := json.Marshal(nconfig)
																		var out bytes.Buffer
																		err := json.Indent(&out, r, "", "\t")
																		if err == nil {
																			ioutil.WriteFile(g.GetRoot()+"/conf/"+"config.json", []byte(out.String()), 0644)
																		}
																		sql := ""
																		listpreout := []string{}
																		lrows, lerr := db.Query("SELECT name FROM [sqlite_master] where type='table' and name like '%pinglog%'")
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
																		//log.Print(sql)
																		g.DLock.Lock()
																		db.Exec(sql)
																		g.DLock.Unlock()
																	}
																} else {
																	preout["info"] = "Refresh illegal!(>0)"
																}
															} else {
																preout["info"] = "Symbol Size illegal!(>0)"
															}

														} else {
															preout["info"] = "Line Thickness  illegal!(>0)"
														}
													} else {
														preout["info"] = "Archive Days !(>0)"
													}
												} else {
													preout["info"] = "Occur Times illegal!(>0)"
												}
											} else {
												preout["info"] = "Average Delay illegal!(>0)"
											}
										} else {
											preout["info"] = "Loss Percent illegal!(<=0 and <=100)"
										}
									} else {
										preout["info"] = "Check Period illegal!(>0)"
									}
								} else {
									preout["info"] = "Timeout illegal!(>0)"
								}
							} else {
								preout["info"] = "Agent Ip illegal!"
							}
						} else {
							preout["info"] = "Agent Name illegal!"
						}

					} else {
						preout["info"] = "decode json error!" + err.Error()
					}
				} else {
					preout["info"] = "param error!"
				}
			} else {
				preout["info"] = "password error!"
			}
		} else {
			preout["info"] = "password empty!"
		}
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
	})
}
