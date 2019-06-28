package funcs

import (
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
)

func StartCloudMonitor() {
	seelog.Info("[func:StartCloudMonitor] ", "starting run StartCloudMonitor ")
	g.CLock.Lock()
	_, err := g.SaveCloudConfig(g.Cfg.Mode["Endpoint"])
	g.CLock.Unlock()
	if err != nil {
		seelog.Error("[func:StartCloudMonitor] Cloud Monitor Error", err)
		return
	}
	g.CLock.Lock()
	g.Cfg.Mode["Status"] = "true"
	g.CLock.Unlock()
	saveerr := g.SaveConfig()
	if saveerr != nil {
		seelog.Error("[func:StartCloudMonitor] Save Cloud Config Error", err)
		return
	}
	seelog.Info("[func:StartCloudMonitor] ", "StartCloudMonitor finish ")

}
