$(function () {
	// we use an inline data source in the example, usually data would
	// be fetched from a server
	var data = [], totalPoints = 200;
	function getRandomData() {
		if (data.length > 0)
			data = data.slice(1);

		while (data.length < totalPoints) {
			var prev = data.length > 0 ? data[data.length - 1] : 50;
			var y = prev + Math.random() * 10 - 5;
			if (y < 0)
				y = 0;
			if (y > 100)
				y = 100;
			data.push(y);
		}

		var res = [];
		for (var i = 0; i < data.length; ++i)
			res.push([i, data[i]])
		return res;
	}

	// setup plot
	var options = {
		yaxis: { min: 0, max: 100 },
		xaxis: { min: 0, max: 100 },
		colors: ["#F90", "#222", "#666", "#BBB"],
		series: {
				   lines: { 
						lineWidth: 2, 
						fill: true,
						fillColor: { colors: [ { opacity: 0.6 }, { opacity: 0.2 } ] },
						steps: false

					}
			   }
	};
	
	var plot = $.plot($("#area-chart"), [ getRandomData() ], options);
});