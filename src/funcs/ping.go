package funcs

import (
	"bufio"
	"github.com/gy-games-libs/mahonia"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type PingSt struct {
	SendPk   string
	RevcPk   string
	LossPk   string
	MinDelay string
	AvgDelay string
	MaxDelay string
}

func pingLinux(Addr string) PingSt {
	var ps PingSt
	cmd := exec.Command("ping", "-w", "60", "-i", "3", "-c", "20", Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	ps.MinDelay = "0"
	ps.AvgDelay = "0"
	ps.MaxDelay = "0"
	ps.SendPk = "20"
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

func pingWindows(Addr string) PingSt {
	var ps PingSt
	var fps PingSt
	fps.SendPk = "20"
	fps.RevcPk = "0"
	fps.MaxDelay = "0"
	fps.MinDelay = "3000"
	fps.AvgDelay = "0"
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		cmd := exec.Command("ping", "-w", "3000", "-n", "1", Addr)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			log.Println(err)
		}
		cmd.Start()
		reader := bufio.NewReader(stdout)
		ps.RevcPk = "0"
		ps.MaxDelay = "0"
		//ps.MinDelay = "3000"
		ps.AvgDelay = "0"
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
			}
			if strings.Contains(line, "最短") {
				packge := strings.Fields(line)
				ps.MinDelay = strings.Split(strings.Split(packge[2], "，")[0], "ms")[0]
				ps.AvgDelay = strings.Split(packge[6], "ms")[0]
				ps.MaxDelay = strings.Split(strings.Split(packge[4], "，")[0], "ms")[0]
			}
		}
		cmd.Wait()

		Delay, _ := strconv.Atoi(ps.MaxDelay)

		GMinDelay, _ := strconv.Atoi(fps.MinDelay)
		if GMinDelay > Delay {
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
		if (stop.Sub(start).Nanoseconds() / 1000000) < 3000 {
			time.Sleep(time.Duration(3000-int(stop.Sub(start).Nanoseconds()/1000000)) * time.Millisecond)
		}
	}
	GRevcPk, _ := strconv.Atoi(fps.RevcPk)
	fps.LossPk = strconv.Itoa(((20 - GRevcPk) / 20) * 100)
	GAvgDelay, _ := strconv.Atoi(fps.AvgDelay)
	fps.AvgDelay = strconv.Itoa(GAvgDelay / 20)
	return fps
}

func Ping(Addr string) PingSt {
	var rt PingSt
	switch os := runtime.GOOS; os {
	case "linux":
		rt = pingLinux(Addr)
	case "windows":
		rt = pingWindows(Addr)
	default:
		log.Fatalf("Unsupported OS type: %s.  Can't establish ping cmd args.\n", os)
	}
	return rt
}
