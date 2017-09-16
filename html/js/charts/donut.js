$(function () {

	var data = [];
	var series = 3;
	for( var i = 0; i<series; i++)
	{
		data[i] = { label: "Series "+(i+1), data: Math.floor(Math.random()*100)+1 }
	}

	$.plot($("#donut-chart"), data,
	{
		colors: ["#19bc9c", "#3398db", "#fad231", "#9b59b6"],
	        series: {
	            pie: { 
	                innerRadius: 0.5,
	                show: true
	            }
	        }
	});
	
});