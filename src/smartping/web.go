package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"encoding/json"
	"strconv"
	"database/sql"
	"time"
	"smartping/mn"
	"github.com/go-resty/resty"
	"bytes"
	"os"
)

// Init of the Web Page template.
var index = template.Must(template.New("index.tpl").Delims("<%", "%>").Funcs(template.FuncMap{"compare": compare,"timestr":timestr}).Parse(`
<% $Showtype := .Showtype%>
<% $LocalIp := .Localip%>
<% $Localname := .Localname%>
<html>
	<head>
	    <title>[<% $Localname %>] SMARTPING</title>
        <link rel="stylesheet" href="https://cdn.bootcss.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
        <link rel="stylesheet" href="https://cdn.bootcss.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
        <script src="https://cdn.bootcss.com/jquery/3.2.0/jquery.min.js"></script>
        <script src="https://cdn.bootcss.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
        <script src="https://cdn.bootcss.com/echarts/3.4.0/echarts.common.min.js"></script>
        <script language="JavaScript" type="text/javascript" src="http://www.my97.net/dp/My97DatePicker/WdatePicker.js"></script>
	</head>
	<body>
		<div id="main" style="margin: auto">
          <div class="row" style="text-align:center">
            <h1>[<% $Localname %>] SMARTPING</h1>
            <p>Total number : <strong><% .State|len %></strong></p>
          </div>
          <br/>
          <div class="row">
              <div class="col-md-1"  ></div>
              <div class="col-md-8"  >
                <div style="text-align:center">
                  <a href="/?t=out" class="label label-<% if eq $Showtype "out" %>success<% else %>default<% end %>">OUT</a>
			      <a href="/?t=in" class="label label-<% if eq $Showtype "in" %>success<% else %>default<% end %>">IN</a>
			      <a target="_blank" href="/api/config.json" class="label label-default" style="float:left">Config</a>
			      <a target="_blank" href="/topology" class="label label-warning" style="float:right">ALERT</a>
		         </div><br/>
              	<%range .State%>
              	   <% if or (eq .Target.Type "CS") (eq $Showtype "out")  %>
                		<div id="<% .Target.Name %>_pannel" style="float:left;width:400px;height:150px;margin-right:10px;" class="showcharts" value="<% .Target.Addr %>" ></div>
                        <% end %>
                	<%end%>
              </div>
              <div class="col-md-2"  >
              <table  class="table table-striped">
					<tr>
						<th>Name</th>
						<th>Addr</th>
						<th>Update</th>
					</tr>
                    	<%range .State%>
                    		<% if eq .Target.Type "CS" %>
					<tr>
						<td><a href="http://<% .Target.Addr %>:8899" ><% .Target.Name %></a></td>
						<td><% .Target.Addr %></td>
						<td><% .LastCheck |timestr %></td>
					</tr>
				<% end %>
                    <%end%>
              </table>

              </div>
              <div class="col-md-1"  ></div>
            </div>
        </div>
        <hr/>
        <div style="text-align:center">SmartPing v<% .Conf.Ver %></div>
        <br/>
    </div>
    <div id="charts" class="modal fade" tabindex="-1" role="dialog">
      <input type="hidden" id="pannelurl" value=""/>
      <div class="modal-dialog" role="document">
        <div class="modal-content"  style="width: 800px;height:550px;">
         <div style="text-align:center;padding-top:10px;">STARTTIME:<input id="starttime" type='text'  onClick="WdatePicker({dateFmt:'yyyy-MM-dd HH:mm'})" > ENDTIME:<input id="endtime" type='text'  onClick="WdatePicker({dateFmt:'yyyy-MM-dd HH:mm'})" > <input class="sgraph" type="button" value="SUBMIT" /></div>
          <div class="modal-body" id="pannel-show" style="width: 800px;height:500px;">
          </div>
        </div><!-- /.modal-content -->
      </div><!-- /.modal-dialog -->
    </div><!-- /.modal -->
    <script>
	$(".sgraph").click(function(){
            start=$("#starttime").val();
            endtime=$("#endtime").val();
            //url='http://'+name+':8899/api/status.json?ip=<% $LocalIp %>'
            url=$("#pannelurl").attr("value")
            console.log(url+"&starttime="+start+"&endtime="+endtime)
            $.get(url+"&starttime="+start+"&endtime="+endtime).done(function (data) {
            var data = JSON.parse(data);
		    myChart.setOption({
			xAxis: {
			    data: data.lastcheck
			},
			series: [{
			    name: 'maxDelay',
			    data: data.maxdelay
			},
			{
			    name: 'minDelay',
			    data: data.mindelay
			},
			{
			    name: 'avgDelay',
			    data: data.avgdelay
			},
			{
			    name: 'lossPk',
			    data: data.losspk
			}]
		    });
        	});
    	});
       opt={
		    title: {
			text: ''
		    },
		    legend: {
			data:['maxDelay','avgDelay','minDelay','lossPk'],
			selected: {
			    'maxDelay' : false,
			    'minDelay' : false
			}
		    },
		    tooltip: {},
		    xAxis: {
			data: []
		    },
		    dataZoom: [{}],
		    yAxis: [{
			type: 'value',
			name: 'Delay',
			position: 'left'
		    }, {
			type: 'value',
			name: 'Package(LOSS)',
			min: 0,
			max: 100,
			position: 'right',
			axisLabel: {
			    formatter: '{value} %'
			}
		    }],
		    series: [{
			    name: 'maxDelay',
			    type: 'line',
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    },
			    data: []
			},
			{
			    name: 'minDelay',
			    type: 'line',
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    },
			    data: []
			},
			{
			    name: 'avgDelay',
			    type: 'line',
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#00CC66'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			},
			{
			    name: 'lossPk',
			    type: 'line',
			    yAxisIndex: 1,
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#FF0000'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			}]
		}
        optmini={
		   title:{
			text: '',
        		'x':'center'
		    },
		    tooltip: {},
		    grid: {
			    left: '3%',
			    right: '3%',
			    bottom: '3%',
			    top: '3%',
			    containLabel: true
		    },
		    xAxis: {
			data: []
		    },
		    yAxis: [{
			type: 'value',
			position: 'left'
		    }, {
			type: 'value',
			min: 0,
			max: 100,
			position: 'right'
		    }],
		    series: [
			{
			    name: 'avgDelay',
			    type: 'line',
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#00CC66'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			},
			{
			    name: 'lossPk',
			    type: 'line',
			    yAxisIndex: 1,
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#FF0000'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			}]
		}
        <%range .State%>
	   <% if or (eq .Target.Type "CS") (eq $Showtype "out")  %>
            var <% .Target.Name %> = echarts.init(document.getElementById('<% .Target.Name %>_pannel'));
            <% .Target.Name %>.setOption(optmini);
            <% .Target.Name %>.setOption({title:{text:'<% if eq $Showtype "out" %><% $Localname %>-><% .Target.Name %><% else %><% .Target.Name %>-><% $Localname %><% end %>'}});
            $.get('<% if eq $Showtype "out" %>/api/status.json?ip=<% .Target.Addr %><% else %>http://<% .Target.Addr %>:8899/api/status.json?ip=<% $LocalIp %><% end %>').done(function (data) {
                var data = JSON.parse(data);
                <% .Target.Name %>.setOption({
                    xAxis: {
                        data: data.lastcheck
                    },
                    series: [{
                        name: 'avgDelay',
                        data: data.avgdelay
                    },
                    {
                        name: 'lossPk',
                        data: data.losspk
                    }]
                    });
            });
          <%end%>
        <%end%>
        var myChart = echarts.init(document.getElementById('pannel-show'));
	     myChart.setOption(opt);
    	    $(".showcharts").click(function(){
            $('#charts').modal('show');
            name=$(this).attr("value")
            <% if eq $Showtype "out" %>
            		url='/api/status.json?ip='+name
            <% else %>
            		url='http://'+name+':8899/api/status.json?ip=<% $LocalIp %>'
            <% end %>
            $("#pannelurl").attr("value",url)
            $.get(url).done(function (data) {
            var data = JSON.parse(data);
		    myChart.setOption({
			xAxis: {
			    data: data.lastcheck
			},
			series: [{
			    name: 'maxDelay',
			    data: data.maxdelay
			},
			{
			    name: 'minDelay',
			    data: data.mindelay
			},
			{
			    name: 'avgDelay',
			    data: data.avgdelay
			},
			{
			    name: 'lossPk',
			    data: data.losspk
			}]
		    });
        });
    });
    </script>
</body>
</html>
`))
var topology = template.Must(template.New("topology.tpl").Delims("<%", "%>").Funcs(template.FuncMap{"json": json.Marshal}).Parse(`
<!DOCTYPE html>
<html style="height: 100%">
   <head>
        <meta charset="utf-8">
        <title>SMARTPING TOPOLOGY</title>
        <link rel="stylesheet" href="https://cdn.bootcss.com/bootstrap/3.3.7/css/bootstrap.min.css" integrity="sha384-BVYiiSIFeK1dGmJRAkycuHAHRg32OmUcww7on3RYdg4Va+PmSTsz/K68vbdEjh4u" crossorigin="anonymous">
        <link rel="stylesheet" href="https://cdn.bootcss.com/bootstrap/3.3.7/css/bootstrap-theme.min.css" integrity="sha384-rHyoN1iRsVXV4nD0JutlnGaslCJuC7uwjduW9SVrLvRYooPp2bWYgmgJQIXwl/Sp" crossorigin="anonymous">
        <script src="https://cdn.bootcss.com/jquery/3.2.0/jquery.min.js"></script>
        <script src="https://cdn.bootcss.com/bootstrap/3.3.7/js/bootstrap.min.js" integrity="sha384-Tc5IQib027qvyjSMfHjOMaLkfuWVxZxUPnCJA7l2mCWNIpG9mGCD8wGNIcPD7Txa" crossorigin="anonymous"></script>
        <script type="text/javascript" src="http://echarts.baidu.com/gallery/vendors/echarts/echarts-all-3.js"></script>
        <script type="text/javascript" src="http://echarts.baidu.com/gallery/vendors/echarts/extension/dataTool.min.js"></script>
        <script type="text/javascript" src="http://echarts.baidu.com/gallery/vendors/echarts/map/js/china.js"></script>
        <script type="text/javascript" src="http://echarts.baidu.com/gallery/vendors/echarts/map/js/world.js"></script>
        <script type="text/javascript" src="http://api.map.baidu.com/api?v=2.0&ak=ZUONbpqGBsYGXNIYHicvbAbM"></script>
        <script type="text/javascript" src="http://echarts.baidu.com/gallery/vendors/echarts/extension/bmap.min.js"></script>
   </head>
   <body style="height: 100%; margin: 0">

       <div class="row" style="height: 100%">
          <div class="col-md-9" style="height: 100%"><div id="container" style="height: 100%"></div></div>
          <div class="col-md-3">
          <% if ne .Alert "" %><audio style="display:none"  autoplay="autoplay"  controls="controls" loop="loop"><source src='<% .Alert %>' type='audio/mp3'  /></audio><% end %>
          <%range .AGraph%>
            <div id="<% index . "From" %><% index . "To" %>_pannel" style="float:left;width:400px;height:150px;margin-right:10px;" class="well"></div>
          <%end%>
          </div>
       </div>
      <script type="text/javascript">
        var dom = document.getElementById("container");
        var myChart = echarts.init(dom);
        var app = {};
        option = null;
	 dataarea = [];
	 <%range .Nlist %>
	     dataarea.push({
		name: '<% index . "name" %>',
		itemStyle: {
			normal: {
				color: "<% index . "color"  %>"
			}
	     	}
	     });
         <%end%>
	  dataline = [];
         <%range .Tlist %>
	  	dataline.push({
				source: '<% index .From "name" %>',
				target: '<% index .To "name" %>',
				lineStyle: {
					normal: {curveness: <% if eq (index .Color) "green" %>0<% else %>0.3<% end %>,color: "<% index .Color %>"}
				}
			})
	  <%end%>
         option = {
            title: {
                text: 'SMARTPING TOPOLOGY'
            },
            tooltip: {},
            animationDurationUpdate: 1500,
            animationEasingUpdate: 'quinticInOut',
            series : [
                {
                    type: 'graph',
                    layout: 'circular',

                    symbolSize: <% .Tsymbolsize %>,
                    focusNodeAdjacency:true,
                    roam: true,
                    label: {
                        normal: {
                            show: true
                        }
                    },
                    edgeSymbol: ['circle', 'arrow'],
                    edgeSymbolSize: [3, 15],
                    edgeLabel: {
                        normal: {
                            textStyle: {
                                fontSize: 15
                            }
                        }
                    },
                    data: dataarea,
                    // links: [],
                    links: dataline,
                    lineStyle: {
                        normal: {
                            opacity: 0.9,
                            width: <% .Tline %>,
                            curveness: 0
                        }
                    }
                }
            ]
        };;
        if (option && typeof option === "object") {
            myChart.setOption(option, true);
        }

      optmini={
		   title:{
			text: '',
        		'x':'center'
		    },
		    tooltip: {},
		    grid: {
			    left: '3%',
			    right: '3%',
			    bottom: '3%',
			    top: '3%',
			    containLabel: true
		    },
		    xAxis: {
			data: []
		    },
		    yAxis: [{
			type: 'value',
			position: 'left'
		    }, {
			type: 'value',
			min: 0,
			max: 100,
			position: 'right'
		    }],
		    series: [
			{
			    name: 'avgDelay',
			    type: 'line',
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#00CC66'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			},
			{
			    name: 'lossPk',
			    type: 'line',
			    yAxisIndex: 1,
			    data: [],
			    itemStyle: {
				normal: {
				    color : '#FF0000'
				}
			    },
			    animation: false,
			    areaStyle: {
				normal: {}
			    },
			    lineStyle: {
				normal: {
				    width: 0
				}
			    }
			}]
		}
        <%range .AGraph%>
            var <% index . "From" %><% index . "To" %> = echarts.init(document.getElementById('<% index . "From" %><% index . "To" %>_pannel'));
            <% index . "From" %><% index . "To" %>.setOption(optmini);
            <% index . "From" %><% index . "To" %>.setOption({title:{text:'<% index . "From" %>-><% index . "To" %>'}});
            $.get('<% index . "Gapi" %>').done(function (data) {
                var data = JSON.parse(data);
                <% index . "From" %><% index . "To" %>.setOption({
                    xAxis: {
                        data: data.lastcheck
                    },
                    series: [{
                        name: 'avgDelay',
                        data: data.avgdelay
                    },
                    {
                        name: 'lossPk',
                        data: data.losspk
                    }]
                    });
            });
        <%end%>
       </script>
   </body>
</html>

`))

func startHttp(port int, state *State ,db *sql.DB ,config Config) {

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
		lock.Lock()
		rows, _ := db.Query("SELECT * FROM pinglog where 1=1 and lastcheck between '"+timeStartStr+"' and '"+timeEndStr+"' "+where+"")
		//foreach all data
		for rows.Next() {
			l := new(LogInfo)
			err := rows.Scan(&l.logtime, &l.ip, &l.name, &l.maxdelay, &l.mindelay, &l.avgdelay, &l.sendpk, &l.revcpk, &l.losspk, &l.lastcheck,)
			if err != nil {
				fmt.Println(err)
			}
			for n, v := range lastcheck{
				if v==l.lastcheck{
					maxdelay[n] = l.maxdelay
					mindelay[n] = l.mindelay
					avgdelay[n] = l.avgdelay
					losspk[n] = l.losspk
				}
			}
		}
		lock.Unlock()
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
		//if loss lager than 30 or avf delay lager than 200 during last 15 min
		//sec,_:=strconv.Atoi(config.Thdchecksec)
		//timeStart := time.Now().Unix()-int64(sec)
		//timeStartStr := time.Unix(timeStart, 0).Format("2006-01-02 15:04")
		var Thdavgdelay string
		var Thdloss string
		var Thdoccnum string
		var timeStart int64
		var timeStartStr string
		for _,v := range state.State{
			lock.Lock()
			sec,_:=strconv.Atoi(config.Thdchecksec)
			timeStart   = time.Now().Unix()-int64(sec)
			Thdloss     = config.Thdloss
			Thdavgdelay = config.Thdavgdelay
			Thdoccnum   = config.Thdoccnum
			for _,t:=range config.Targets {
				if t.Name==v.Target.Name{
					if t.Thdchecksec !=""{
						sec,_:=strconv.Atoi(t.Thdchecksec)
						timeStart = time.Now().Unix()-int64(sec)
					}
					if t.Thdloss != ""{
						Thdloss = t.Thdloss
					}
					if t.Thdavgdelay != ""{
						Thdavgdelay = t.Thdavgdelay
					}
					if t.Thdoccnum != ""{
						Thdoccnum = t.Thdoccnum
					}
				}
			}
			timeStartStr = time.Unix(timeStart, 0).Format("2006-01-02 15:04")
			rows, _ := db.Query("SELECT * FROM pinglog where 1=1 and lastcheck > '"+timeStartStr+"' and ip= '"+v.Target.Addr+"'")
			//log.Println("SELECT * FROM pinglog where 1=1 and lastcheck > '"+timeStartStr+"' and ip= '"+v.Target.Addr+"'")
			preout[v.Target.Name]="true"
			var showtimes int
			for rows.Next() {
				l := new(LogInfo)
				rows.Scan(&l.logtime, &l.ip, &l.name, &l.maxdelay, &l.mindelay, &l.avgdelay, &l.sendpk, &l.revcpk, &l.losspk, &l.lastcheck,)
				lp,_:=strconv.Atoi(l.losspk)
				ad,_:=strconv.Atoi(l.avgdelay)
				thlp,_:=strconv.Atoi(Thdloss)
				thad,_:=strconv.Atoi(Thdavgdelay)
				if(lp>thlp || ad>thad){
					showtimes = showtimes+1
				}
			}
			oct,_:=strconv.Atoi(Thdoccnum)
			if showtimes>=oct {
				preout[v.Target.Name]="false"
			}
			lock.Unlock()
		}
		out, _ := json.Marshal(preout)
		fmt.Fprintln(w, string(out))
	})

	//Topology Alert
	http.HandleFunc("/topology", func(w http.ResponseWriter, r *http.Request) {
		var alertgraph  []map[string]string
		randinfo   := make(map[int]map[string]string)
		i :=0
		var randint  []int
		sl := new(showlist)
		sl.Alert = ""
		linestatus := make(map[string]string)
		for _,v := range config.Targets{
			var st string
			st = "green"
			randinfo[i] = map[string]string{
				"name":v.Name,
				"ip":v.Addr,
			}
			randint=append(randint,i)
			i = i+1
			if v.Type=="CS"{
				resp, _ := resty.SetTimeout(time.Second).R().Get("http://"+v.Addr+":8899/api/topology.json")
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
			}
			tostatus := map[string]string{
				"name"   : v.Name,
				"type"   : v.Type,
				"color" : st,
			}
			sl.Nlist = append(sl.Nlist,tostatus)
			sl.AGraph = alertgraph
		}
		//set line color
		sl.Status = linestatus
		zuheres := mn.Zuhe2(i,randint)
		//Get Full Arrangement
		for _,rd :=range []string{"FROM","TO"}{
			for _,v :=range  zuheres{
				tt := new(topo)
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

	//Index
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state.Lock.Lock()
		defer state.Lock.Unlock()
		r.ParseForm()
		state.Showtype="out"
		if len(r.Form["t"]) > 0 {
			state.Showtype = r.Form["t"][0]
		}
		state.Conf = config
		//log.Println(state)
		err := index.Execute(w, state)
		if err != nil {
			log.Println("ERR:",err)
		}
	})
	s := fmt.Sprintf(":%d", port)
	log.Println("starting to listen on ", s)
	log.Printf("Get status on http://localhost%s/status", s)
	err := http.ListenAndServe(s, nil)
	if err != nil {
		log.Println("ERR:",err)
	}
	log.Println("Server on 8899 stopped")
	os.Exit(0)
}
