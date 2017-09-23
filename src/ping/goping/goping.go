package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gy-games-libs/go-fastping"
	"net"
	"strconv"
	"time"
	"os"
)

type Str struct {
	Flag    bool
	Timeout string
	Message string
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "ip address!")
	debug := flag.Bool("d", false, "debug!")
	flag.Parse()
	Ping(*ip,*debug)
}

func Ping(Addr string,debug bool) {
	var rt Str
	rt.Message = "timeout"
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", Addr)
	if err == nil {
		p.MaxRTT = time.Millisecond * 3000
		p.AddIPAddr(ra)
		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
			rt.Flag = true
			rt.Message=""
			rt.Timeout = strconv.Itoa(int(rtt.Nanoseconds()/1000000))
			if debug == true{
				fmt.Println("[func:Ping] Addr:",Addr," Delay:",rt.Timeout)
			}
			out, _ := json.Marshal(rt)
			fmt.Print(string(out))
			os.Exit(0)
		}
		err = p.Run()
		if err != nil {
			rt.Flag = false
			rt.Message = err.Error()
		}
	} else {
		rt.Flag = false
		rt.Message = err.Error()
	}
	out, _ := json.Marshal(rt)
	fmt.Print(string(out))

}