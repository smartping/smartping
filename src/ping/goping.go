package ping

import (
	"../g"
	"bufio"
	"encoding/json"
	"github.com/gy-games-libs/seelog"
	"io"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

type Str struct {
	Flag    bool
	Timeout string
	Message string
}

func GoPing(Addr string) g.PingSt {

	var ping string
	switch os := runtime.GOOS; os {
	case "windows":
		ping = g.GetRoot() + "/bin/goping.exe"
	default:
		ping = g.GetRoot() + "/bin/goping"
	}
	var MaxDelay,MinDelay,AllDelay,Delay float64
	SendPK := 0
	RevcPK := 0
	MaxDelay = 0
	MinDelay = -1
	AllDelay = 0
	RevcBool := false
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		RevcBool = false
		SendPK = SendPK + 1
		cmd := exec.Command(ping, "-ip", Addr)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			seelog.Error("[func:GoPing]", Addr, " Ping Command Error", err)
			break
		}
		cmd.Start()
		reader := bufio.NewReader(stdout)
		Delay = 0
		for {
			var rt Str
			l, err2 := reader.ReadString('\n')
			err = json.Unmarshal([]byte(l), &rt)
			if err != nil {
				seelog.Error("[func:GoPing] JsonUnmarshal", err)
				break
			}
			if rt.Flag == true {
				Delay, _ = strconv.ParseFloat(rt.Timeout,64)
				if Delay == 0{
					SendPK = SendPK-1
				}else{
					RevcPK = RevcPK + 1
					RevcBool = true
					if MinDelay == -1 || MinDelay > Delay {
						MinDelay = Delay
					}
					if MaxDelay < Delay {
						MaxDelay = Delay
					}
					AllDelay = AllDelay + Delay
				}
				break
			} else {
				if rt.Message != "timeout" {
					seelog.Error("[func:GoPing] Ping Error", rt.Message)
				}
				break
			}
			if err2 != nil || io.EOF == err2 {
				break
			}
		}
		cmd.Wait()
		stop := time.Now()
		seelog.Debug("[func:GoPing] Addr:", Addr, " Cnt:", ic, " CurrentStatus:", RevcBool, " CurrentDelay:", Delay, " Send:", SendPK, " Revc:", RevcPK, " MaxDelay:", MaxDelay, " MinDelay:", MinDelay, " SMCost:", stop.Sub(start))
		if (stop.Sub(start).Nanoseconds() / 1000000) < 3000 {
			during := time.Duration(3000-int(stop.Sub(start).Nanoseconds()/1000000)) * time.Millisecond
			seelog.Debug("[func:GoPing]", Addr, " Gorouting Sleep.", during)
			time.Sleep(during)
		}

	}
	var fps g.PingSt
	fps.MaxDelay = strconv.FormatFloat(MaxDelay, 'f', 3, 64)
	if MinDelay == -1 {
		fps.MinDelay = "0"
	} else {
		fps.MinDelay = strconv.FormatFloat(MinDelay, 'f', 3, 64)
	}
	if AllDelay > 0 {
		fps.AvgDelay = strconv.FormatFloat(AllDelay / float64(RevcPK), 'f', 3, 64)
	}
	fps.SendPk = strconv.Itoa(SendPK)
	fps.RevcPk = strconv.Itoa(RevcPK)
	fps.LossPk = strconv.Itoa(((SendPK - RevcPK) / SendPK) * 100)
	seelog.Info("[func:GoPing] Finish Addr:", Addr, " MaxDelay:", fps.MaxDelay, " MinDelay:", fps.MinDelay, " AvgDelay:", fps.AvgDelay, " Send:", fps.SendPk, " Revc:", fps.RevcPk, " LossPK:", fps.LossPk)
	return fps
}
