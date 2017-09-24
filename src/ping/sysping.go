package ping

import (
	"../g"
	"bufio"
	"github.com/gy-games-libs/seelog"
	"io"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func SysPing(Addr string) g.PingSt {
	var args [5]string
	switch os := runtime.GOOS; os {
	case "windows":
		args[0] = "-n"
		args[1] = "1"
		args[2] = "-w"
		args[3] = "3000"
	default:
		args[0] = "-c"
		args[1] = "1"
		args[2] = "-w"
		args[3] = "3"
	}
	args[4] = Addr
	SendPK := 0
	RevcPK := 0
	MaxDelay := 0
	MinDelay := -1
	AllDelay := 0
	RevcBool := false
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		RevcBool = false
		SendPK = SendPK + 1
		cmd := exec.Command("ping", args[0:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			seelog.Error("[func:SysPing]", Addr, " Ping Command Error", err)
			break
		}
		cmd.Start()
		reader := bufio.NewReader(stdout)
		Delay := 0
	ploop:
		for {
			l, err2 := reader.ReadString('\n')
			if strings.Contains(l, Addr) && strings.Contains(l, "ms") {
				re := regexp.MustCompile(`([\d.]*\s*)ms`)
				ms := re.FindAllStringSubmatch(l, -1)
				if len(ms) > 0 && len(ms[0]) == 2 {
					DelayF64, _ := strconv.ParseFloat(strings.Replace(ms[0][1], " ", "", -1), 32)
					Delay = int(DelayF64)
					RevcPK = RevcPK + 1
					RevcBool = true
					if MinDelay == -1 || MinDelay > Delay {
						MinDelay = Delay
					}
					if MaxDelay < Delay {
						MaxDelay = Delay
					}
					AllDelay = AllDelay + Delay
					break ploop
				}
			}
			if err2 != nil || io.EOF == err2 {
				break ploop
			}
		}
		cmd.Wait()
		stop := time.Now()
		seelog.Debug("[func:SysPing] Addr:", Addr, " Cnt:", ic, " CurrentStatus:", RevcBool, " CurrentDelay:", Delay, " Send:", SendPK, " Revc:", RevcPK, " MaxDelay:", MaxDelay, " MinDelay:", MinDelay, " SMCost:", stop.Sub(start))
		if (stop.Sub(start).Nanoseconds() / 1000000) < 3000 {
			during := time.Duration(3000-int(stop.Sub(start).Nanoseconds()/1000000)) * time.Millisecond
			seelog.Debug("[func:SysPing]", Addr, " Gorouting Sleep.", during)
			time.Sleep(during)
		}

	}
	var fps g.PingSt
	fps.MaxDelay = strconv.Itoa(MaxDelay)
	if MinDelay == -1 {
		fps.MinDelay = "0"
	} else {
		fps.MinDelay = strconv.Itoa(MinDelay)
	}
	if AllDelay > 0 {
		fps.AvgDelay = strconv.Itoa(AllDelay / RevcPK)
	}
	fps.SendPk = strconv.Itoa(SendPK)
	fps.RevcPk = strconv.Itoa(RevcPK)
	fps.LossPk = strconv.Itoa(((SendPK - RevcPK) / SendPK) * 100)
	seelog.Info("[func:SysPing] Finish Addr:", Addr, " MaxDelay:", fps.MaxDelay, " MinDelay:", fps.MinDelay, " AvgDelay:", fps.AvgDelay, " Revc:", fps.RevcPk, " LossPK:", fps.LossPk)
	return fps
}
