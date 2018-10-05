function Refresh(){
    window.location.reload();
}
function AgentMode(mode,status){
    if(mode=="cloud"){
        if (status==true){
            $("#cloudbrand").html("<i class='icon icon-cloud'></i>&nbsp;");
        }else{
            $("#cloudbrand").html("<i class='icon icon-cloud icon-danger'></i>&nbsp;");
        }
    }
}