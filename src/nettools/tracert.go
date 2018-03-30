package nettools

import (
	"errors"
	"github.com/gy-games-libs/golang/x/net/icmp"
	"github.com/gy-games-libs/golang/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

func RunTrace(Addr string, maxrtt time.Duration, maxttl int, maxtimeout int) ([]ICMP, error) {
	hops := make([]ICMP, 0, maxttl)
	var res pkg
	var err error
	res.dest, err = net.ResolveIPAddr("ip", Addr)
	if err != nil {
		return nil, errors.New("Unable to resolve destination host")
	}
	res.maxrtt = maxrtt
	//res.id = rand.Int() % 0x7fff
	res.id = rand.Intn(65535)
	res.seq = 1
	res.msg = icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: res.id, Seq: res.seq}}
	res.netmsg, err = res.msg.Marshal(nil)
	if nil != err {
		return nil, err
	}
	timeouts := 0
	for i := 1; i <= maxttl; i++ {
		next := res.Send(i)
		next.MaxRTT = next.RTT
		next.MinRTT = next.RTT
		for j := 0; j < 2; j++ {
			tnext := res.Send(i)
			if tnext.RTT >= next.RTT {
				next.MaxRTT = tnext.RTT
			}
			if tnext.MinRTT <= next.RTT {
				next.MinRTT = tnext.RTT
			}
		}
		next.AvgRTT = time.Duration((next.MaxRTT + next.RTT + next.MinRTT) / 3)
		hops = append(hops, next)
		if next.Final {
			break
		}
		if next.Timeout {
			timeouts++
		} else {
			timeouts = 0
		}
		if timeouts == maxtimeout {
			break
		}
	}
	return hops, nil
}
