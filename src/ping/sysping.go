package ping

import (
	"bufio"
	//"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"../g"
	"github.com/gy-games-libs/seelog"
	"regexp"
	//"io"
	"io"
	"runtime"
)



func SysPing(Addr string) g.PingSt {
	var args [5]string
	switch os := runtime.GOOS; os {
	case "windows":
		args[0]="-n"
		args[1]="1"
		args[2]="-w"
		args[3]="3000"
	default:
		args[0]="-c"
		args[1]="1"
		args[2]="-w"
		args[3]="3"
	}
	args[4]=Addr
	var ps g.PingSt
	var fps g.PingSt
	fps.SendPk = "20"
	fps.RevcPk = "0"
	fps.MaxDelay = "0"
	fps.MinDelay = "3000"
	fps.AvgDelay = "0"
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		cmd := exec.Command("ping",args[0:]...)
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
		ploop:
		for {
			l, err2 := reader.ReadString('\n')
			if strings.Contains(l,Addr) && strings.Contains(l,"ms"){
				re := regexp.MustCompile(`([\d.]*\s*)ms`)
				ms := re.FindAllStringSubmatch(l,-1)
				if len(ms)>0 && len(ms[0])==2{
					delay = ms[0][1]
					ps.RevcPk = "1"
					break ploop
				}
			}
			if err2 != nil || io.EOF == err2 {
				break ploop
			}
		}
		cmd.Wait()
		//if cmd.Process.Pid>0{
		//	cmd.Process.Kill()
		//}
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

func gosysping(reader *bufio.Reader, Addr string, ch chan string){

	//cmd := exec.Command("ping", Addr)
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	seelog.Error("[func:CmdPing]",Addr," Ping Command Error",err)
	//}
	//cmd.Start()
	//reader := bufio.NewReader(stdout)
	for {
		l, _ := reader.ReadString('\n')
		if strings.Contains(l,Addr) && strings.Contains(l,"ms"){
			//ch <-l
			re := regexp.MustCompile(`([\d.]*\s*)ms`)
			ms := re.FindAllStringSubmatch(l,-1)
			if len(ms)>0 && len(ms[0])==2{
				delay := ms[0][1]
				//ps.RevcPk = "1"
				ch <- delay
				break
			}
		}
	}

}