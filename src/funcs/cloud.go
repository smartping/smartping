package funcs

import (
	"github.com/cihub/seelog"
	"github.com/smartping/smartping/src/g"
	"time"
)

func StartCloudMonitor(cnt int) {
	if cnt < 3 {
		seelog.Info("[func:StartCloudMonitor] ", "starting run StartCloudMonitor ")
		_, err := g.SaveCloudConfig(g.Cfg.Cendpoint, true)
		if err != nil {
			seelog.Error("[func:StartCloudMonitor] Cloud Monitor Error", err)
			g.Cfg.Cstatus = false
			StartCloudMonitor(cnt + 1)
			return
		}
		saveerr := g.SaveConfig()
		if saveerr != nil {
			seelog.Error("[func:StartCloudMonitor] Save Cloud Config Error", err)
			g.Cfg.Cstatus = false
			StartCloudMonitor(cnt + 1)
			return
		}
		g.Cfg.Cstatus = true
		seelog.Info("[func:StartCloudMonitor] ", "StartCloudMonitor finish ")
		time.Sleep(1 * time.Second)
	}
}
