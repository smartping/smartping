package funcs

import (
	"bufio"
	"encoding/json"
	"github.com/elves-project/agent/src/g"
	"github.com/gy-games-libs/seelog"
	"io"
	"os/exec"
	"runtime"
)

type PingSt struct {
	SendPk   string
	RevcPk   string
	LossPk   string
	MinDelay string
	AvgDelay string
	MaxDelay string
}

func Ping(Addr string) PingSt {
	var rt PingSt
	var ping string
	switch os := runtime.GOOS; os {
	case "windows":
		ping = g.GetRoot() + "/bin/ping.exe"
	default:
		ping = g.GetRoot() + "/bin/ping"
	}
	cmd := exec.Command(ping, "-ip", Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		seelog.Error(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		seelog.Debug(line)
		err = json.Unmarshal([]byte(line), &rt)
		if err != nil {
			seelog.Error("[func:Ping] ", err)
		}
		if err2 != nil || io.EOF == err2 {
			break
		}
	}
	cmd.Wait()
	seelog.Debug("[func:Ping] Finnal", Addr, " MaxDelay:"+rt.MaxDelay+" MinDelay:"+rt.MinDelay+" AvgDelay:"+rt.AvgDelay+" SendPK:"+rt.SendPk+" RevcPk:"+rt.RevcPk+" LossPK:"+rt.LossPk)
	return rt
}
