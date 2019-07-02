function Refresh(){
    window.location.reload();
}
function AgentMode(Mode){
    if(Mode["Type"]=="cloud"){
        $("#cloudbrand").html("<i class='icon icon-cloud'></i>&nbsp;");
        $("#cfgUrl").attr("href","config.html?cloud")
        $("#banner_last_ck_time").html("最后同步时间："+Mode["LastSuccTime"])
    }
}