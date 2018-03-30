package nettools

import (
	"encoding/binary"
	"github.com/gy-games-libs/golang/x/net/icmp"
	"github.com/gy-games-libs/golang/x/net/ipv4"
	"net"
	"time"
)

type pkg struct {
	conn     net.PacketConn
	ipv4conn *ipv4.PacketConn
	msg      icmp.Message
	netmsg   []byte
	id       int
	seq      int
	maxrtt   time.Duration
	dest     net.Addr
}

type ICMP struct {
	Addr    net.Addr
	RTT     time.Duration
	MaxRTT  time.Duration
	MinRTT  time.Duration
	AvgRTT  time.Duration
	Final   bool
	Timeout bool
	Down    bool
	Error   error
}

func (t *pkg) Send(ttl int) ICMP {
	var hop ICMP
	var err error
	t.conn, hop.Error = net.ListenPacket("ip4:icmp", "0.0.0.0")
	if nil != err {
		return hop
	}
	defer t.conn.Close()
	t.ipv4conn = ipv4.NewPacketConn(t.conn)
	defer t.ipv4conn.Close()
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
		switch result.Type {
		case ipv4.ICMPTypeEchoReply:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id == rply.ID && t.seq == rply.Seq {
					hop.Final = true
					hop.RTT = time.Since(sendOn)
					return hop
				}

			}
		case ipv4.ICMPTypeTimeExceeded:
			if rply, ok := result.Body.(*icmp.TimeExceeded); ok {
				if len(rply.Data) > 24 {
					if uint16(t.id) == binary.BigEndian.Uint16(rply.Data[24:26]) {
						hop.RTT = time.Since(sendOn)
						return hop
					}
				}
			}
		case ipv4.ICMPTypeDestinationUnreachable:
			if rply, ok := result.Body.(*icmp.Echo); ok {
				if t.id == rply.ID && t.seq == rply.Seq {
					hop.Down = true
					hop.RTT = time.Since(sendOn)
					return hop
				}

			}
		}
	}
}
