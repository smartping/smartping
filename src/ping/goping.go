package ping

import (
	"runtime"
	"os/exec"
	"github.com/gy-games-libs/seelog"
	"encoding/json"
	"io"
	"bufio"
	"../g"
	"strconv"
	"time"
)

type Str struct {
	Flag    bool
	Timeout string
	Message string
}

func GoPing(Addr string) g.PingSt {
	var rt Str
	var ps g.PingSt
	var fps g.PingSt
	var ping string
	switch os := runtime.GOOS; os {
	case "windows":
		ping = g.GetRoot() + "/bin/goping.exe"
	default:
		ping = g.GetRoot() + "/bin/goping"
	}
	fps.SendPk = "20"
	fps.RevcPk = "0"
	fps.MaxDelay = "0"
	fps.MinDelay = "3000"
	fps.AvgDelay = "0"
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		cmd := exec.Command(ping, "-ip", Addr)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			seelog.Error("[func:CmdPing]",Addr," Ping Command Error",err)
			break
		}
		cmd.Start()
		reader := bufio.NewReader(stdout)
		ps.RevcPk   = "0"
		ps.MaxDelay = "0"
		ps.MinDelay = "0"
		ps.AvgDelay = "0"
		delay := "0"
		//ploop:
		for {
			l, err2 := reader.ReadString('\n')
			err = json.Unmarshal([]byte(l), &rt)
			if err != nil {
				seelog.Error("[func:SPing] JsonUnmarshal", err)
				break
			}
			if rt.Flag ==true{
				delay = rt.Timeout
				ps.RevcPk = "1"
				seelog.Debug("[func:SPing] ",rt)
				break
			}else{
				if rt.Message !="timeout"{
					seelog.Error("[func:SPing] Ping Error", rt.Message)
				}
				break
			}
			if err2 != nil || io.EOF == err2 {
				break
			}
		}
		cmd.Wait()

		Delay, _ := strconv.Atoi(delay)

		GMinDelay, _ := strconv.Atoi(fps.MinDelay)
		if Delay>0 && GMinDelay > Delay {
			fps.MinDelay = strconv.Itoa(Delay)
		}

		GMaxDelay, _ := strconv.Atoi(fps.MaxDelay)
		if GMaxDelay < Delay {
			fps.MaxDelay = strconv.Itoa(Delay)
		}

		GAvgDelay, _ := strconv.Atoi(fps.AvgDelay)
		fps.AvgDelay = strconv.Itoa(GAvgDelay + Delay)

		if ps.RevcPk == "1" {
			GRevcPk, _ := strconv.Atoi(fps.RevcPk)
			fps.RevcPk = strconv.Itoa(GRevcPk + 1)
		}
		stop := time.Now()
		seelog.Debug("[func:CmdPing] Addr:",Addr," Cnt:",ic," Current:",delay," Revc:",fps.RevcPk," MaxDelay:",fps.MaxDelay," MinDelay:",fps.MinDelay," SMCost:",stop.Sub(start))
		if (stop.Sub(start).Nanoseconds() / 1000000) < 3000 {
			during := time.Duration(3000-int(stop.Sub(start).Nanoseconds()/1000000)) * time.Millisecond
			seelog.Debug("[func:CmdPing]",Addr," Gorouting Sleep.",during)
			time.Sleep(during)
		}

	}
	if fps.MinDelay=="3000"{
		fps.MinDelay = "0"
	}
	GRevcPk, _ := strconv.Atoi(fps.RevcPk)
	fps.LossPk = strconv.Itoa(((20 - GRevcPk) / 20) * 100)
	GAvgDelay, _ := strconv.Atoi(fps.AvgDelay)
	if(GRevcPk>0){
		fps.AvgDelay = strconv.Itoa(GAvgDelay / GRevcPk)
	}
	seelog.Info("[func:Ping] Finish Addr:",Addr," MaxDelay:",fps.MaxDelay," MinDelay:",fps.MinDelay," AvgDelay:",fps.AvgDelay," Revc:",fps.RevcPk," LossPK:",fps.LossPk)
	return fps
}