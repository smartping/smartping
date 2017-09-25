package funcs

import (
	"bufio"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"path/filepath"
	"regexp"
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
	var pingpath string
	pingfile,err:=PathExists(getCurrentDirectory()+"/ping.exe")
	if pingfile {
		pingpath = getCurrentDirectory()+"/ping.exe"
	}else{
		pingpath = "ping"
	}
	fps.SendPk = "20"
	fps.RevcPk = "0"
	fps.MaxDelay = "0"
	fps.MinDelay = "3000"
	fps.AvgDelay = "0"
	for ic := 0; ic < 20; ic++ {
		start := time.Now()
		cmd := exec.Command(pingpath, "-w", "3000", "-n", "1", Addr)
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
			line, err2 := reader.ReadString('\n')
			if err2 != nil || io.EOF == err2 {
				break
			}
			if strings.Contains(line, "%") {
				re := regexp.MustCompile(`Packets: Sent = (?P<sent>\d{1,9}), Received = (?P<rec>\d{1,9}), Lost = (?P<loss>\d{0,3}) \(\d{1,3}% loss\),`)
				ps.SendPk = re.FindStringSubmatch(line)[1]
				ps.RevcPk = re.FindStringSubmatch(line)[2]
			}
			if strings.Contains(line, "Minimum") {
				re := regexp.MustCompile(`Minimum = (?P<min>\d{1,9})ms, Maximum = (?P<max>\d{1,9})ms, Average = (?P<avg>\d{1,9})ms`)
				ps.MinDelay = re.FindStringSubmatch(line)[1]
				ps.MaxDelay = re.FindStringSubmatch(line)[2]
				ps.AvgDelay = re.FindStringSubmatch(line)[3]
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
	GRevcPk, err := strconv.Atoi(fps.RevcPk)
	if err != nil {
        panic(err)
	}
	LossPkFloat:=(20-float32(GRevcPk))/20
	fps.LossPk = strconv.Itoa((int)(LossPkFloat*100))
	GAvgDelay, _ := strconv.Atoi(fps.AvgDelay)
	fps.AvgDelay = strconv.Itoa(GAvgDelay / 20)
	return fps
}

func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
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
