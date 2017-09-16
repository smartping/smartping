package funcs

import (
	"bufio"
	"github.com/gy-games-libs/mahonia"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

type PingSt struct {
	SendPk   string
	RevcPk   string
	LossPk   string
	MinDelay string
	AvgDelay string
	MaxDelay string
}

func pingLinux(Addr string, cnt string) PingSt {
	var ps PingSt
	cmd := exec.Command("ping", "-w", "5", "-c", cnt, Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	ps.MinDelay = "0"
	ps.AvgDelay = "0"
	ps.MaxDelay = "0"
	ps.SendPk = cnt
	ps.RevcPk = "0"
	ps.LossPk = "100"
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if strings.Contains(line, "packets transmitted") {
			packge := strings.Fields(line)
			ps.SendPk = packge[0]
			ps.RevcPk = packge[3]
			ps.LossPk = strings.Split(packge[5], "%")[0]
		}
		if strings.Contains(line, "rtt min/avg/max/mdev") {
			rrttmp := strings.Fields(line)
			rrt := strings.Split(rrttmp[3], "/")
			ps.MinDelay = rrt[0]
			ps.AvgDelay = rrt[1]
			ps.MaxDelay = rrt[2]
		}
	}
	cmd.Wait()
	return ps
}

func pingWindows(Addr string, cnt string) PingSt {
	var ps PingSt
	cmd := exec.Command("ping", "-w", " 5000", "-n", cnt, Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	ps.MinDelay = "0"
	ps.AvgDelay = "0"
	ps.MaxDelay = "0"
	ps.SendPk = cnt
	ps.RevcPk = "0"
	ps.LossPk = "100"
	for {
		l, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		var line string
		var dec mahonia.Decoder
		dec = mahonia.NewDecoder("gbk")
		line = dec.ConvertString(l)
		if strings.Contains(line, "%") {
			packge := strings.Fields(line)
			ps.SendPk = strings.Split(packge[3], "，")[0]
			ps.RevcPk = strings.Split(packge[5], "，")[0]
			ps.LossPk = strings.Split(strings.Split(packge[8], "(")[1], "%")[0]
			//log.Print(packge)

		}
		if strings.Contains(line, "最短") {
			packge := strings.Fields(line)
			ps.MinDelay = strings.Split(strings.Split(packge[2], "，")[0], "ms")[0]
			ps.AvgDelay = strings.Split(packge[6], "ms")[0]
			ps.MaxDelay = strings.Split(strings.Split(packge[4], "，")[0], "ms")[0]
		}
	}
	cmd.Wait()
	return ps
}

func Ping(Addr string, cnt string) PingSt {
	var rt PingSt
	switch os := runtime.GOOS; os {
	case "linux":
		rt = pingLinux(Addr, cnt)
	case "windows":
		rt = pingWindows(Addr, cnt)
	default:
		log.Fatalf("Unsupported OS type: %s.  Can't establish ping cmd args.\n", os)
	}
	return rt
}
