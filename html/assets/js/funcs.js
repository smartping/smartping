function Refresh(){
    window.location.reload();
}
function AgentMode(mode,status){
    if(mode=="cloud"){
        $(".localmode").remove();
        $(".cloudmode").show();
        if (status==true){
            //$(".cicon").addClass("icon-success");
            $(".cloudmodeonline").show();
            $(".cloudmodeoffline").remove();
        }else{
            $(".cicon").addClass("icon-danger")
            $(".cloudmodeonline").remove();
            $(".cloudmodeoffline").show();
        }
    }else{
        $(".cloudmode").remove();
    }
}