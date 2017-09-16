$(function () {
    var sin = [], cos = [];
    for (var i = 0; i < 10; i += 0.5) {
        sin.push([i, Math.sin(i)]);
        cos.push([i, Math.cos(i)]);
    }

    var plot = $.plot($("#line-chart"),
           [ { data: sin, label: "sin(x)"}, { data: cos, label: "cos(x)" } ], {
               series: {
                   lines: { show: true },
                   points: { show: true }
               },
               
               grid: { hoverable: true, clickable: true },
               yaxis: { min: -1.1, max: 1.1 },
			   xaxis: { min: 0, max: 9 },
    	colors: ["#F90", "#222", "#666", "#BBB"]
             });
});