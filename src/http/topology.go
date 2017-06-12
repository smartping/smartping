package http

import (
	"net/http"
	"strconv"
	"github.com/gy-games-libs/resty"
	"time"
	"encoding/json"
	"log"
	"database/sql"
	"../funcs"
	"../g"
)

func configTopologyRoutes(port int, state *g.State ,db *sql.DB ,config g.Config){

	//Topology Alert
	http.HandleFunc("/topology", func(w http.ResponseWriter, r *http.Request) {
		var alertgraph  []map[string]string
		randinfo   := make(map[int]map[string]string)
		i :=0
		var randint  []int
		sl := new(g.Showlist)
		sl.Alert = ""
		linestatus := make(map[string]string)
		for _,v := range config.Targets{
			var st string
			st = "green"
			if v.Type=="CS"{
				startTime := funcs.CurrentTimeMillis()
				sec,_   :=strconv.Atoi(config.Topotimeout)
				resp, _ := resty.SetTimeout(time.Second*time.Duration(sec)).R().Get("http://"+v.Addr+":8899/api/topology.json")
				if resp.StatusCode()==200{
					map2 := make(map[string]interface{})
					json.Unmarshal([]byte(resp.String()), &map2)
					for f,s := range map2{
						if s=="true"{
							linestatus[v.Name+f]="green"
						}else{
							agraph := map[string]string{
								"From":v.Name,
								"To":f,
								"Gapi":"http://"+v.Addr+":8899/api/status.json?name="+f,
							}
							alertgraph = append(alertgraph,agraph)
							linestatus[v.Name+f]="red"
							sl.Alert=config.Alertsound
						}
					}
				}else{
					st = "red"
					sl.Alert=config.Alertsound
				}
				endTime  := funcs.CurrentTimeMillis()
				//Costtime :=
				randinfo[i] = map[string]string{
					"name":v.Name,
					"costtime":strconv.Itoa(int(endTime - startTime)),
					"ip":v.Addr,
				}
			}else{
				randinfo[i] = map[string]string{
					"name":v.Name,
					"costtime":"0",
					"ip":v.Addr,
				}
			}
			randint=append(randint,i)

			tostatus := map[string]string{
				"name"   	: randinfo[i]["name"],
				"costtime"   	: randinfo[i]["costtime"],
				"type"   	: v.Type,
				"color" 	: st,
			}
			i = i+1
			sl.Nlist = append(sl.Nlist,tostatus)
			sl.AGraph = alertgraph
		}
		//set line color
		sl.Status = linestatus
		zuheres := funcs.Zuhe2(i,randint)
		//Get Full Arrangement
		for _,rd :=range []string{"FROM","TO"}{
			for _,v :=range  zuheres{
				tt := new(g.Topo)
				if rd=="FROM"{
					tt.From = randinfo[v[0]]
					tt.To   = randinfo[v[1]]
				}else{
					tt.From = randinfo[v[1]]
					tt.To   = randinfo[v[0]]
				}
				k := string(tt.From["name"])+string(tt.To["name"])
				if linestatus[k] != ""{
					tt.Color = linestatus[k]
				}else{
					tt.Color = "#FFFF00"
				}
				//Except The Client Line
				for _,ck := range config.Targets{
					if ck.Name == tt.From["name"]{
						if ck.Type=="CS"{
							sl.Tlist = append(sl.Tlist,tt)
						}
					}
				}
			}
		}
		sl.Tline=config.Tline
		sl.Tsymbolsize=config.Tsymbolsize
		err := topology.Execute(w, sl)
		if err != nil {
			log.Println("ERR:",err)
		}
	})

}
