package nettools

import (
	"errors"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"
)

type Mtr struct {
	Host  string
	Send  int
	Loss  int
	Last  time.Duration
	Avg   time.Duration
	Best  time.Duration
	Wrst  time.Duration
	StDev float64
}

func RunMtr(Addr string, maxrtt time.Duration, maxttl int, maxtimeout int) ([]Mtr, error) {
	result := []Mtr{}
	Lock := sync.Mutex{}
	var wg sync.WaitGroup
	mtr := map[int][]ICMP{}
	var err error
	timeouts := 0
	for ttl := 1; ttl <= maxttl; ttl++ {
		id := rand.Intn(65535)
		seq := rand.Intn(65535)
		res := pkg{
			maxrtt: maxrtt,
			id:     id,
			seq:    seq,
			msg:    icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: id, Seq: seq}},
		}
		res.dest, err = net.ResolveIPAddr("ip", Addr)
		if err != nil {
			return result, errors.New("Unable to resolve destination host")
		}
		res.netmsg, err = res.msg.Marshal(nil)
		if nil != err {
			return result, err
		}
		next := res.Send(ttl)
		if next.Timeout {
			timeouts++
		} else {
			timeouts = 0
		}
		if timeouts == maxtimeout {
			break
		}
		Lock.Lock()
		mtr[ttl] = append(mtr[ttl], next)
		Lock.Unlock()
		wg.Add(1)
		go func(ittl int) {
			defer wg.Done()
			for j := 1; j < 10; j++ {
				id := rand.Intn(65535)
				seq := rand.Intn(65535)
				res := pkg{
					maxrtt: maxrtt,
					id:     id,
					seq:    seq,
					msg:    icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: id, Seq: seq}},
				}
				res.dest, err = net.ResolveIPAddr("ip", Addr)
				if err != nil {
					return
				}
				res.netmsg, err = res.msg.Marshal(nil)
				if nil != err {
					return
				}
				nowTime := time.Now()
				next := res.Send(ittl)
				Lock.Lock()
				mtr[ittl] = append(mtr[ittl], next)
				Lock.Unlock()
				time.Sleep(time.Second - time.Now().Sub(nowTime))
			}
		}(ttl)
		if next.Final {
			break
		}
	}
	wg.Wait()
	for i := 1; i <= len(mtr); i++ {
		imtr := Mtr{}
		for id, val := range mtr[i] {
			if val.Addr != nil {
				imtr.Host = val.Addr.String()
			} else {
				if imtr.Host == "" {
					imtr.Host = "???"
				}
			}
			imtr.Send += 1
			if val.Timeout {
				imtr.Loss += 1
			} else {
				if imtr.Wrst < val.RTT {
					imtr.Wrst = val.RTT
				}
				if id == 0 {
					imtr.Best = val.RTT
				}
				if imtr.Best > val.RTT {
					imtr.Best = val.RTT
				}
				imtr.Avg += val.RTT
				imtr.Last = val.RTT
			}
		}
		if (imtr.Send - imtr.Loss) > 0 {
			imtr.Avg = imtr.Avg / time.Duration(imtr.Send-imtr.Loss)
			for _, val := range mtr[i] {
				if !val.Timeout {
					v := (float64(val.RTT.Nanoseconds()) / 1e6) - (float64(imtr.Avg.Nanoseconds()) / 1e6)
					imtr.StDev += v * v
				}
			}
			imtr.StDev = math.Sqrt(imtr.StDev / float64(imtr.Send-imtr.Loss))
		}
		result = append(result, imtr)

	}
	return result, nil
}

/*
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
*/
