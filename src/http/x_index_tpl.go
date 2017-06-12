package http

var indexTemplate = `
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
`
