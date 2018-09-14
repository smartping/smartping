function Refresh(){
    window.location.reload();
}
function AgentMode(mode,status){
    if(mode=="cloud"){
        $(".localmode").remove();
        $(".cloudmode").show();
        if (status==true){
            $(".cloudmodeonline").show();
            $(".cloudmodeoffline").remove();
        }else{
            $(".cloudmodeonline").remove();
            $(".cloudmodeoffline").show();
        }
    }else{
        $(".cloudmode").remove();
    }
}