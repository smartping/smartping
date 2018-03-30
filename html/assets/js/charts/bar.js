$(function () {
	var data = new Array ();
    var ds = new Array();
	
	data.push ([[1,25],[2,34],[3,37],[4,45],[5,56]]);
	data.push ([[1,13],[2,29],[3,25],[4,23],[5,31]]);
	data.push ([[1,8],[2,13],[3,19],[4,15],[5,14]]);
	data.push ([[1,20],[2,43],[3,29],[4,23],[5,25]]);
 
    for (var i=0, j=data.length; i<j; i++) {
    	
	     ds.push({
	        data:data[i],
	        grid:{
            hoverable:true
        },
	        bars: {
	            show: true, 
	            barWidth: 0.2, 
	            order: 1,
	            lineWidth: 0.5, 
				fillColor: { colors: [ { opacity: 0.65 }, { opacity: 1 } ] }
	        }
	    });
	}
	    
    $.plot($("#bar-chart"), ds, {
    	colors: ["#F90", "#222", "#666", "#BBB"]
                

    });
                

    
});