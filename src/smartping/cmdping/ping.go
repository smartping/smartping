package cmdping

import (
	"io"
	"strings"
	"log"
	"os/exec"
	"bufio"
	"runtime"
	"github.com/axgle/mahonia"
)

type PingSt struct{
 	SendPk string
 	RevcPk string
 	LossPk string
 	MinDelay string
 	AvgDelay string
 	MaxDelay string
}

func pingLinux(Addr string , localip string, cnt string) PingSt{
	var ps PingSt
	cmd := exec.Command("ping", "-c", cnt, Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		if(strings.Contains(line, "packets transmitted")){
			packge := strings.Fields(line)
			ps.SendPk = packge[0]
			ps.RevcPk = packge[3]
			ps.LossPk = strings.Split(packge[5],"%")[0]
			//log.Print(t.Addr,":",line)
		}
		if(strings.Contains(line, "rtt min/avg/max/mdev")){
			rrttmp := strings.Fields(line)
			rrt := strings.Split(rrttmp[3],"/")
			ps.MinDelay = rrt[0]
			ps.AvgDelay = rrt[1]
			ps.MaxDelay = rrt[2]
			log.Print("Addr:",Addr," SendPk:",ps.SendPk," RevcPk:",ps.LossPk," LossPk:",ps.LossPk," | MinDelay:",ps.MinDelay," AvgDelay:",ps.AvgDelay," MaxDelay:",ps.MaxDelay)
		}
	}
	cmd.Wait()
	return ps;
}

func pingWindows(Addr string , cnt string ) PingSt{
	var ps PingSt
	cmd := exec.Command("ping", "-n", cnt, Addr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	//实时循环读取输出流中的一行内容
	for {
		l, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		var line string
		var dec mahonia.Decoder
		dec = mahonia.NewDecoder("gbk")
		line = dec.ConvertString(l)
		if(strings.Contains(line, "%")){
			packge := strings.Fields(line)
			ps.SendPk = strings.Split(packge[3],"，")[0]
			ps.RevcPk = strings.Split(packge[5],"，")[0]
			ps.LossPk = strings.Split(strings.Split(packge[8],"(")[1],"%")[0]
		}
		if(strings.Contains(line, "最短")){
			packge := strings.Fields(line)
			ps.MinDelay = strings.Split(strings.Split(packge[2],"，")[0],"ms")[0]
			ps.AvgDelay = strings.Split(strings.Split(packge[4],"，")[0],"ms")[0]
			ps.MaxDelay = strings.Split(packge[6],"ms")[0]
			log.Print("Addr:",Addr," SendPk:",ps.SendPk," RevcPk:",ps.LossPk," LossPk:",ps.LossPk," | MinDelay:",ps.MinDelay," AvgDelay:",ps.AvgDelay," MaxDelay:",ps.MaxDelay)
		}
	}
	cmd.Wait()
	return ps;
}

func Ping(Addr string , localip string , cnt string ) PingSt{
	var rt PingSt
	switch os := runtime.GOOS; os {
	case "linux":
		rt = pingLinux(Addr,localip, cnt)
	case "windows":
		rt = pingWindows(Addr, cnt)
	default:
		log.Fatalf("Unsupported OS type: %s.  Can't establish ping cmd args.\n", os)
	}
	return  rt
}