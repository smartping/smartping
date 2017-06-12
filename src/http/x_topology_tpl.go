package http

var topologyTemplate = `
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
		category:'<% index . "costtime" %>',
		draggable: "true",
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
            tooltip: {
            		show: true
            },
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
                        	  //position: 'right',
                    	   	  //formatter: '{c}',
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

`
