package nettools

import (
	"encoding/binary"
	"errors"
	"github.com/gy-games-libs/golang/x/net/icmp"
	"github.com/gy-games-libs/golang/x/net/ipv4"
	"math/rand"
	"net"
	"time"
)

type trace struct {
	conn     net.PacketConn
	ipv4conn *ipv4.PacketConn
	msg      icmp.Message
	netmsg   []byte
	id       int
	maxrtt   time.Duration
	maxttl   int
	dest     net.Addr
}

type Hop struct {
	Addr    net.Addr
	Host    string
	RTT     time.Duration
	MaxRTT  time.Duration
	MinRTT  time.Duration
	AvgRTT  time.Duration
	Final   bool
	Timeout bool
	Down    bool
	Error   error
}

func RunTrace(host string, maxrtt time.Duration, maxttl int, maxtimeout int) ([]Hop, error) {
	hops := make([]Hop, 0, maxttl)
	var res trace
	var err error
	addrList, err := net.LookupIP(host)
	if nil != err {
		return nil, err
	}
	for _, addr := range addrList {
		if addr.To4() != nil {
			res.dest, err = net.ResolveIPAddr("ip4:icmp", addr.String())
			break
		}
	}
	if nil == res.dest {
		return nil, errors.New("Unable to resolve destination host")
	}
	res.maxrtt = maxrtt
	res.maxttl = maxttl
	res.id = rand.Int() % 0x7fff
	res.msg = icmp.Message{Type: ipv4.ICMPTypeEcho, Code: 0, Body: &icmp.Echo{ID: res.id, Seq: 1}}
	res.netmsg, err = res.msg.Marshal(nil)
	if nil != err {
		return nil, err
	}
	res.conn, err = net.ListenPacket("ip4:icmp", "0.0.0.0")
	if nil != err {
		return nil, err
	}
	defer res.conn.Close()
	res.ipv4conn = ipv4.NewPacketConn(res.conn)
	defer res.ipv4conn.Close()
	timeouts := 0
	for i := 1; i <= maxttl; i++ {
		next := res.Step(i)
		next.MaxRTT = next.RTT
		next.MinRTT = next.RTT
		for j := 0; j < 2; j++ {
			tnext := res.Step(i)
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

// Step sends one echo packet and waits for result
func (t *trace) Step(ttl int) Hop {
	var hop Hop
	hop.Error = t.conn.SetReadDeadline(time.Now().Add(t.maxrtt))
	if nil != hop.Error {
		return hop
	}
	if nil != t.ipv4conn {
		hop.Error = t.ipv4conn.SetTTL(ttl)
	}
	if nil != hop.Error {
		return hop
	}
	sendOn := time.Now()
	if nil != t.ipv4conn {
		_, hop.Error = t.conn.WriteTo(t.netmsg, t.dest)
	}
	if nil != hop.Error {
		return hop
	}
	buf := make([]byte, 1500)
	for {
		var readLen int
		readLen, hop.Addr, hop.Error = t.conn.ReadFrom(buf)
		if nerr, ok := hop.Error.(net.Error); ok && nerr.Timeout() {
			hop.Timeout = true
			return hop
		}
		if nil != hop.Error {
			return hop
		}
		var result *icmp.Message
		if nil != t.ipv4conn {
			result, hop.Error = icmp.ParseMessage(1, buf[:readLen])
		}
		if nil != hop.Error {
			return hop
		}
		hop.RTT = time.Since(sendOn)
		switch result.Type {
		case ipv4.ICMPTypeEchoReply:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id != rply.ID {
					continue
				}
				hop.Final = true
				return hop
			}
		case ipv4.ICMPTypeTimeExceeded:
			if rply, ok := result.Body.(*icmp.TimeExceeded); ok {
				if len(rply.Data) > 24 {
					if uint16(t.id) != binary.BigEndian.Uint16(rply.Data[24:26]) {
						continue
					}
					return hop
				}
			}
		case ipv4.ICMPTypeDestinationUnreachable:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id != rply.ID {
					continue
				}
				hop.Down = true
				return hop
			}
		}
	}
}
